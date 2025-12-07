package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/parser"
	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

func main() {
	// Command line flags
	port := flag.Int("port", 6379, "Port number to listen on")
	replicaof := flag.String("replicaof", "", "flag to start a Redis server as a replica")
	dbDir := flag.String("dir", "", "the path to the directory where the RDB file is stored")
	dbFileName := flag.String("dbfilename", "", "the name of the RDB file")
	flag.Parse()

	// Data stores
	kvStore := make(map[string]types.KVEntry)
	listStore := make(map[string][]string)
	streamStore := make(map[string][]types.StreamEntry)
	clients := make(map[net.Conn]*types.Client)
	replicaConnections := &types.ReplicaConns{}
	blockingClients := make(map[string][]chan any)
	listMutex := &sync.Mutex{}
	streamsMutex := &sync.Mutex{}
	masterReplicationID := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	masterReplOffset := 0
	masterReplOffsetMutex := &sync.Mutex{}
	acks := make(chan int, 10)

	defaultUser := &types.User{
		Name:      "default",
		Flags:     []string{"nopass"},
		Passwords: []string{},
	}
	userMutex := &sync.Mutex{}

	if *dbDir != "" && *dbFileName != "" {
		rdbData, err := parser.ParseRDB(*dbDir + "/" + *dbFileName)
		if err == nil {
			now := time.Now()
			for _, v := range rdbData.Keys {
				kv := types.KVEntry{
					Value:     v.Value,
					CreatedAt: now,
				}

				if v.HasExpire {
					switch v.Type {
					case "s":
						// v.ExpireAt is absolute time in seconds
						// We need to calculate how many seconds from now until expiration
						secondsUntilExpire := v.ExpireAt - now.Unix()
						if secondsUntilExpire > 0 {
							kv.Ex = int(secondsUntilExpire)
						} else {
							// Already expired, skip this key
							continue
						}
					case "ms":
						// v.ExpireAt is absolute time in milliseconds
						// Calculate milliseconds from now until expiration
						msUntilExpire := v.ExpireAt - (now.UnixNano() / 1e6)
						if msUntilExpire > 0 {
							kv.Px = int(msUntilExpire)
						} else {
							// Already expired, skip this key
							continue
						}
					}
				}
				kvStore[v.Key] = kv
			}
		}
	}

	// If the server is started as a replica, connect to the master
	if *replicaof != "" {
		masterConn, err := utils.ConnectToMaster(*replicaof, *port)
		if err != nil {
			fmt.Println("Failed to connect to master:", err)
			os.Exit(1)
		}
		clients[masterConn] = &types.Client{}
		if slices.Contains(defaultUser.Flags, "nopass") {
			clients[masterConn].Connected = true
		} else {
			clients[masterConn].Connected = false
		}
		go handleConnection(masterConn, kvStore, listStore, streamStore, clients, replicaConnections, listMutex, streamsMutex, blockingClients, *replicaof, masterReplicationID, true, &masterReplOffset, masterReplOffsetMutex, acks, *dbDir, *dbFileName, defaultUser, userMutex)
	}

	// Start listening for connections on the specified port
	add := "0.0.0.0:" + strconv.Itoa(*port)
	l, err := net.Listen("tcp", add)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer func() {
		if err := l.Close(); err != nil {
			fmt.Printf("Error closing listener: %v\n", err)
		}
	}()
	fmt.Printf("Server is listening on port %d\n", *port)

	// Main loop to accept new connections
	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		clients[con] = &types.Client{}
		if slices.Contains(defaultUser.Flags, "nopass") {
			clients[con].Connected = true
		} else {
			clients[con].Connected = false
		}
		go handleConnection(con, kvStore, listStore, streamStore, clients, replicaConnections, listMutex, streamsMutex, blockingClients, *replicaof, masterReplicationID, false, &masterReplOffset, masterReplOffsetMutex, acks, *dbDir, *dbFileName, defaultUser, userMutex)
	}
}

// handleConnection manages a single client connection.
func handleConnection(con net.Conn, kvStore map[string]types.KVEntry, listStore map[string][]string, streamStore map[string][]types.StreamEntry, clients map[net.Conn]*types.Client, replicaConnections *types.ReplicaConns, listMutex *sync.Mutex, streamsMutex *sync.Mutex, blockingClients map[string][]chan any, replicaof string, masterReplicationID string, isReplica bool, masterReplOffset *int, masterReplOffsetMutex *sync.Mutex, acks chan int, dbDir, dbFileName string, defaultUser *types.User, userMutex *sync.Mutex) {

	defer func() {
		if err := con.Close(); err != nil {
			fmt.Printf("Error closing connection: %v\n", err)
		}
	}()
	reader := bufio.NewReader(con)
	client := clients[con]
	replicaOffset := 0

	// If this is a replica connection, perform the initial handshake.
	if isReplica {
		// Consume the "+FULLRESYNC..." response from the master.
		_, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("replica: error reading FULLRESYNC line:", err)
			return
		}

		// Consume the RDB file sent by the master.
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("replica: error reading RDB bulk string header:", err)
			return
		}

		if line[0] == '$' {
			length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
			if err != nil {
				fmt.Println("replica: error parsing RDB length:", err)
				return
			}
			if length > 0 {
				// Discard the RDB file content.
				n, err := reader.Discard(length)
				if err != nil {
					fmt.Println("replica: error discarding RDB content:", err)
					return
				}
				if n != length {
					fmt.Printf("replica: mismatched RDB discard. expected %d, got %d\n", length, n)
					return
				}
			}
		}
	}

	// Loop to read and process commands from the client.
	for {
		// Parse the RESP array from the client's request.
		commands, bytesRead, err := parser.ParseRespArray(reader)
		if err != nil {
			fmt.Println("Error parsing:", err.Error())
			return
		}
		if len(commands) == 0 {
			continue
		}

		if !clients[con].Connected && strings.ToUpper(commands[0]) != "AUTH" {
			_, err := con.Write([]byte("-NOAUTH Authentication required.\r\n"))
			if err != nil {
				fmt.Printf("Error NOAUTH response: %v\n", err)
			}
			continue
		}

		// If in a MULTI block, queue commands until EXEC or DISCARD.
		if client.InMulti && (strings.ToUpper(commands[0]) != "EXEC" && strings.ToUpper(commands[0]) != "DISCARD") {
			client.Commands = append(client.Commands, commands)
			if _, err := con.Write([]byte("+QUEUED\r\n")); err != nil {
				fmt.Printf("Error writing QUEUED response: %v\n", err)
				return
			}
			continue
		}
		var response string
		var commandBytes []byte

		isWriteCommand := false
		switch strings.ToUpper(commands[0]) {
		case "SET", "INCR", "RPUSH", "LPUSH", "LPOP", "XADD":
			isWriteCommand = true
		}

		if isWriteCommand {
			commandBytes = utils.GetCommandBytes(commands)
			masterReplOffsetMutex.Lock()
			*masterReplOffset += len(commandBytes)
			masterReplOffsetMutex.Unlock()

			replicaConnections.Lock()
			for _, conn := range replicaConnections.Conns {
				if conn != nil {
					if _, err := conn.Write(commandBytes); err != nil {
						fmt.Printf("Error writing command to replica: %v\n", err)
					}
				}
			}
			replicaConnections.Unlock()
		}

		// Dispatch the command to the appropriate handler.
		switch strings.ToUpper(commands[0]) {
		case "PING":
			response = handlers.Ping()
		case "INFO":
			response = handlers.Info(commands, replicaof, masterReplicationID)
		case "ECHO":
			response = handlers.Echo(commands)
		case "SET":
			response = handlers.Set(commands, kvStore)
		case "INCR":
			response = handlers.Incr(commands, kvStore)
		case "GET":
			response = handlers.Get(commands, kvStore)
		case "RPUSH":
			response = handlers.Rpush(commands, listStore, listMutex, blockingClients)
		case "LPUSH":
			response = handlers.Lpush(commands, listStore, listMutex)
		case "LRANGE":
			response = handlers.Lrange(commands, listStore, listMutex)
		case "LLEN":
			response = handlers.Llen(commands, listStore, listMutex)
		case "LPOP":
			response = handlers.Lpop(commands, listStore, listMutex)
		case "BLPOP":
			response = handlers.Blpop(commands, listStore, listMutex, blockingClients)
		case "TYPE":
			response = handlers.Type(commands, kvStore, listStore, streamStore)
		case "XADD":
			response = handlers.Xadd(commands, streamStore, streamsMutex, blockingClients)
		case "XRANGE":
			response = handlers.Xrange(commands, streamStore, streamsMutex)
		case "XREAD":
			response = handlers.Xread(commands, streamStore, streamsMutex, blockingClients)
		case "MULTI":
			response = handlers.Multi(client)
		case "EXEC":
			if !client.InMulti {
				response = "-ERR EXEC without MULTI\r\n"
			} else if len(client.Commands) > 0 {
				for _, cmd := range client.Commands {
					isWriteCommand := false
					switch strings.ToUpper(cmd[0]) {
					case "SET", "INCR", "RPUSH", "LPUSH", "LPOP", "XADD":
						isWriteCommand = true
					}
					if isWriteCommand {
						commandBytes = utils.GetCommandBytes(cmd)
						masterReplOffsetMutex.Lock()
						*masterReplOffset += len(commandBytes)
						masterReplOffsetMutex.Unlock()

						replicaConnections.Lock()
						for _, conn := range replicaConnections.Conns {
							if conn != nil {
								if _, err := conn.Write(commandBytes); err != nil {
									fmt.Printf("Error writing command to replica: %v\n", err)
								}
							}
						}
						replicaConnections.Unlock()
					}
				}
				handlers.Exec(con, clients, kvStore, listStore, streamStore, listMutex, streamsMutex, blockingClients)
			} else {
				response = "*0\r\n"
				client.InMulti = false
			}
		case "DISCARD":
			if !client.InMulti {
				response = "-ERR DISCARD without MULTI\r\n"
			} else {
				client.InMulti = false
				client.Commands = nil
				response = "+OK\r\n"
			}
		case "REPLCONF":
			response = handlers.Replconf(con, commands, replicaOffset)
		case "PSYNC":
			handlers.Psync(con, masterReplicationID, replicaConnections)
		case "WAIT":
			response = handlers.Wait(commands, replicaConnections, acks, *masterReplOffset)
		case "KEYS":
			response = handlers.Keys(commands, dbDir, dbFileName)
		case "ACL":
			response = handlers.Acl(commands, defaultUser, userMutex)
		case "AUTH":
			client := clients[con]
			response = handlers.Auth(commands, *defaultUser, client)

		default:
			response = "-ERR unknown command\r\n"
		}

		// Track the replica's offset for replication purposes.
		if isReplica {
			replicaOffset += bytesRead
		}
		// Send the response back to the client.
		if response != "" {
			// For replicas, only respond to REPLCONF commands.
			if !isReplica || (isReplica && commands[0] == "REPLCONF") {
				if _, err := con.Write([]byte(response)); err != nil {
					fmt.Printf("Error writing response: %v\n", err)
					return
				}
			}
		}
	}
}

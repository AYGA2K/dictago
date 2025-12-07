package handlers

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

// Exec executes all previously queued commands in a transaction.
func Exec(con net.Conn, clients map[net.Conn]*types.Client, kvStore map[string]types.KVEntry, listStore map[string][]string, streamStore map[string][]types.StreamEntry, listMutex, streamsMutex *sync.Mutex, blockingClients map[string][]chan any) {
	client := clients[con]
	// Reset the MULTI state for the client.
	defer func() {
		client.InMulti = false
		client.Commands = nil
	}()

	// If there are no queued commands, return an empty array.
	if len(client.Commands) == 0 {
		if _, err := con.Write([]byte("*0\r\n")); err != nil {
			fmt.Printf("Error writing empty EXEC response: %v\n", err)
		}
		return
	}

	// Create a slice to store the responses of each command.
	resp := []string{}
	// Iterate through the queued commands and execute them.
	for _, commands := range client.Commands {
		switch strings.ToUpper(commands[0]) {
		case "PING":
			resp = append(resp, Ping())
		case "ECHO":
			resp = append(resp, Echo(commands))
		case "SET":
			resp = append(resp, Set(commands, kvStore))
		case "INCR":
			resp = append(resp, Incr(commands, kvStore))
		case "GET":
			resp = append(resp, Get(commands, kvStore))
		case "RPUSH":
			resp = append(resp, Rpush(commands, listStore, listMutex, blockingClients))
		case "LPUSH":
			resp = append(resp, Lpush(commands, listStore, listMutex))
		case "LRANGE":
			resp = append(resp, Lrange(commands, listStore, listMutex))
		case "LLEN":
			resp = append(resp, Llen(commands, listStore, listMutex))
		case "LPOP":
			resp = append(resp, Lpop(commands, listStore, listMutex))
		case "BLPOP":
			resp = append(resp, Blpop(commands, listStore, listMutex, blockingClients))
		case "TYPE":
			resp = append(resp, Type(commands, kvStore, listStore, streamStore))
		case "XADD":
			resp = append(resp, Xadd(commands, streamStore, streamsMutex, blockingClients))
		case "XRANGE":
			resp = append(resp, Xrange(commands, streamStore, streamsMutex))
		case "XREAD":
			resp = append(resp, Xread(commands, streamStore, streamsMutex, blockingClients))
		}
	}

	// If there are responses, format them as a RESP array and send them to the client.
	if len(resp) > 0 {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("*%d\r\n", len(resp)))
		for _, r := range resp {
			builder.WriteString(r)
		}
		if _, err := con.Write([]byte(builder.String())); err != nil {
			fmt.Printf("Error writing EXEC response: %v\n", err)
		}
	}
}

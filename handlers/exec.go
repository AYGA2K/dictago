package handlers

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

func Exec(con net.Conn, clients map[net.Conn]*types.Client, m map[string]types.SetArg, mlist map[string][]string, streams map[string][]types.StreamEntry, listMutex, streamsMutex *sync.Mutex, waitingClients map[string][]chan any, replicas *types.ReplicaConns) {
	client := clients[con]
	client.InMulti = false
	numCommands := len(client.Commands)
	responses := make([]string, numCommands)
	for i, cmd := range client.Commands {
		var response string
		switch strings.ToUpper(cmd[0]) {
		case "PING":
			response = Ping()
		case "ECHO":
			response = Echo(cmd)
		case "SET":
			response = Set(cmd, m, replicas)
		case "INCR":
			response = Incr(cmd, m, replicas)
		case "GET":
			response = Get(cmd, m)
		case "RPUSH":
			response = Rpush(cmd, mlist, listMutex, waitingClients, replicas)
		case "LPUSH":
			response = Lpush(cmd, mlist, listMutex, replicas)
		case "LRANGE":
			response = Lrange(cmd, mlist, listMutex)
		case "LLEN":
			response = Llen(cmd, mlist, listMutex)
		case "LPOP":
			response = Lpop(cmd, mlist, listMutex, replicas)
		case "BLPOP":
			response = Blpop(cmd, mlist, listMutex, waitingClients)
		case "TYPE":
			response = Type(cmd, m, mlist, streams)
		case "XADD":
			response = Xadd(cmd, streams, streamsMutex, waitingClients, replicas)
		case "XRANGE":
			response = Xrange(cmd, streams, streamsMutex)
		case "XREAD":
			response = Xread(cmd, streams, streamsMutex, waitingClients)
		default:
			response = "-ERR unknown command in MULTI\r\n"
		}
		responses[i] = response
	}
	client.Commands = nil
	responseArray := fmt.Sprintf("*%d\r\n", numCommands)
	for _, res := range responses {
		responseArray += res
	}
	fmt.Fprint(con, responseArray)
}

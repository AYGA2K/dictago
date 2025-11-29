package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

// It appends one or multiple values to a list.
func Rpush(commands []string, m map[string][]string, listMutex *sync.Mutex, waitingClients map[string][]chan any, replicas *types.ReplicaConns) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'rpush' command\r\n"
	}
	listMutex.Lock()
	defer listMutex.Unlock()
	key := commands[1]
	// If the key does not exist, create a new list.
	if val, ok := m[key]; !ok {
		list := []string{}
		list = append(list, commands[2:]...)
		m[commands[1]] = list

		// If there are clients waiting on this key (e.g., from BLPOP), notify one of them.
		if clients, ok := waitingClients[key]; ok && len(clients) > 0 {
			clientChan := clients[0]
			waitingClients[key] = clients[1:]
			close(clientChan)
		}

		// Propagate the RPUSH command to all replicas.
		replicas.Lock()
		defer replicas.Unlock()
		for _, conn := range replicas.Conns {
			if conn != nil {
				command := fmt.Sprintf("*%d\r\n", len(commands))
				for _, part := range commands {
					command += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
				}
				conn.Write([]byte(command))
			}
		}

		// Return the length of the list after the push operation.
		return fmt.Sprintf(":%d\r\n", len(list))
	} else {
		// If the key exists, append the new values to the existing list.
		val = append(val, commands[2:]...)
		m[key] = val

		// If there are clients waiting on this key, notify one of them.
		if clients, ok := waitingClients[key]; ok && len(clients) > 0 {
			clientChan := clients[0]
			waitingClients[key] = clients[1:]
			close(clientChan)
		}

		// Propagate the RPUSH command to all replicas.
		replicas.Lock()
		defer replicas.Unlock()
		for _, conn := range replicas.Conns {
			if conn != nil {
				command := fmt.Sprintf("*%d\r\n", len(commands))
				for _, part := range commands {
					command += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
				}
				conn.Write([]byte(command))
			}
		}
		// Return the length of the list after the push operation.
		return fmt.Sprintf(":%d\r\n", len(val))
	}
}

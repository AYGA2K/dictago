package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

func Rpush(commands []string, m map[string][]string, listMutex *sync.Mutex, waitingClients map[string][]chan any, replicas *types.ReplicaConns) string {
	if len(commands) > 2 {
		listMutex.Lock()
		defer listMutex.Unlock()
		key := commands[1]
		if val, ok := m[key]; !ok {
			list := []string{}
			list = append(list, commands[2:]...)
			m[commands[1]] = list
			if clients, ok := waitingClients[key]; ok && len(clients) > 0 {
				clientChan := clients[0]
				waitingClients[key] = clients[1:]
				close(clientChan)
			}
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
			return fmt.Sprintf(":%d\r\n", len(list))
		} else {
			val = append(val, commands[2:]...)
			m[key] = val
			if clients, ok := waitingClients[key]; ok && len(clients) > 0 {
				clientChan := clients[0]
				waitingClients[key] = clients[1:]
				close(clientChan)
			}
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
			return fmt.Sprintf(":%d\r\n", len(val))
		}
	}
	return ""
}

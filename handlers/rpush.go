package handlers

import (
	"fmt"
	"sync"
)

// It appends one or multiple values to a list.
func Rpush(commands []string, listStore map[string][]string, listMutex *sync.Mutex, blockingClients map[string][]chan any) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'rpush' command\r\n"
	}
	listMutex.Lock()
	defer listMutex.Unlock()
	key := commands[1]
	// If the key does not exist, create a new list.
	if val, ok := listStore[key]; !ok {
		list := []string{}
		list = append(list, commands[2:]...)
		listStore[commands[1]] = list

		// If there are clients waiting on this key (e.g., from BLPOP), notify one of them.
		if clients, ok := blockingClients[key]; ok && len(clients) > 0 {
			clientChan := clients[0]
			blockingClients[key] = clients[1:]
			close(clientChan)
		}

		// Return the length of the list after the push operation.
		return fmt.Sprintf(":%d\r\n", len(list))
	} else {
		// If the key exists, append the new values to the existing list.
		val = append(val, commands[2:]...)
		listStore[key] = val

		// If there are clients waiting on this key, notify one of them.
		if clients, ok := blockingClients[key]; ok && len(clients) > 0 {
			clientChan := clients[0]
			blockingClients[key] = clients[1:]
			close(clientChan)
		}

		// Return the length of the list after the push operation.
		return fmt.Sprintf(":%d\r\n", len(val))
	}
}

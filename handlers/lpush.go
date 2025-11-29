package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

// It prepends one or multiple values to a list.
func Lpush(commands []string, m map[string][]string, listMutex *sync.Mutex, replicas *types.ReplicaConns) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'lpush' command\r\n"
	}
	listMutex.Lock()
	defer listMutex.Unlock()
	key := commands[1]
	valuesToAdd := commands[2:]
	// Reverse the slice of values to add, so they are prepended in the correct order.
	utils.ReverseSlice(valuesToAdd)
	list := []string{}
	list = append(list, valuesToAdd...)
	// If the key does not exist, create a new list.
	if val, ok := m[key]; !ok {
		m[key] = list
	} else {
		// If the key exists, prepend the new values to the existing list.
		list = append(list, val...)
		m[key] = list
	}

	// Propagate the LPUSH command to all replicas.
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
}

package handlers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

// It removes and returns the first element of a list.
func Lpop(commands []string, m map[string][]string, listMutex *sync.Mutex, replicas *types.ReplicaConns) string {
	listMutex.Lock()
	defer listMutex.Unlock()
	if len(commands) < 2 {
		return "-ERR wrong number of arguments for 'lpop' command\r\n"
	}

	key := commands[1]
	val, ok := m[key]
	// If the key does not exist or the list is empty, return null.
	if !ok || len(val) == 0 {
		return "$-1\r\n"
	}

	// Propagate the LPOP command to all replicas.
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

	// If a count is provided, pop that many elements.
	if len(commands) == 3 {
		count, err := strconv.Atoi(commands[2])
		if err != nil || count <= 0 {
			return "-ERR value is not an integer or out of range\r\n"
		}

		if count > len(val) {
			count = len(val)
		}

		popped := val[:count]
		m[key] = val[count:]

		// Return the popped elements as an array of bulk strings.
		response := fmt.Sprintf("*%d\r\n", len(popped))
		for _, v := range popped {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
		}
		return response
	}

	// If no count is provided, pop one element.
	popped := val[0]
	m[key] = val[1:]
	// Return the popped element as a bulk string.
	return fmt.Sprintf("$%d\r\n%s\r\n", len(popped), popped)
}

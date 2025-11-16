package handlers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

func Lpop(commands []string, m map[string][]string, listMutex *sync.Mutex, replicas *types.ReplicaConns) string {
	listMutex.Lock()
	defer listMutex.Unlock()
	if len(commands) < 2 {
		return "wrong number of arguments for 'lpop' command\r\n"
	}

	key := commands[1]
	val, ok := m[key]
	if !ok || len(val) == 0 {
		return "$-1\r\n"
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

	if len(commands) == 3 {
		count, err := strconv.Atoi(commands[2])
		if err != nil || count <= 0 {
			return "value is not an integer or out of range\r\n"
		}

		if count > len(val) {
			count = len(val)
		}

		popped := val[:count]
		m[key] = val[count:]

		response := fmt.Sprintf("*%d\r\n", len(popped))
		for _, v := range popped {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
		}
		return response
	}

	popped := val[0]
	m[key] = val[1:]
	return fmt.Sprintf("$%d\r\n%s\r\n", len(popped), popped)
}

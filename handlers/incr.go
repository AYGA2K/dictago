package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AYGA2K/dictago/types"
)

func Incr(commands []string, m map[string]types.SetArg, replicas *types.ReplicaConns) string {
	if len(commands) > 1 {
		key := commands[1]
		if val, ok := m[key]; ok {
			intVal, err := strconv.Atoi(val.Value)
			if err != nil {
				return "-ERR value is not an integer or out of range\r\n"
			}
			intVal++
			val.Value = strconv.Itoa(intVal)
			m[key] = val
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
			return fmt.Sprintf(":%d\r\n", intVal)
		}
		new := types.SetArg{
			Value:     "1",
			CreatedAt: time.Now(),
		}
		m[key] = new
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
		return ":1\r\n"
	}
	return ""
}

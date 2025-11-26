package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AYGA2K/dictago/types"
)

func Set(commands []string, m map[string]types.SetArg, replicas *types.ReplicaConns) string {
	if len(commands) > 2 {
		setArg := types.SetArg{
			Value:     commands[2],
			CreatedAt: time.Now(),
		}
		if len(commands) > 4 {
			switch commands[3] {
			case "EX":
				ex, err := strconv.Atoi(commands[4])
				if err != nil {
					return ""
				}
				setArg.Ex = ex
			case "PX":
				px, err := strconv.Atoi(commands[4])
				if err != nil {
					return ""
				}
				setArg.Px = px
			}
		}
		m[commands[1]] = setArg
		replicas.Lock()
		defer replicas.Unlock()
		fmt.Println("SET: ", setArg)
		for _, conn := range replicas.Conns {
			if conn != nil {
				command := fmt.Sprintf("*%d\r\n", len(commands))
				for _, part := range commands {
					command += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
				}
				conn.Write([]byte(command))
			}
		}
		return "+OK\r\n"
	}
	return ""
}

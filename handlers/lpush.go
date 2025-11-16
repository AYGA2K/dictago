package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

func Lpush(commands []string, m map[string][]string, listMutex *sync.Mutex, replicas *types.ReplicaConns) string {
	if len(commands) > 2 {
		listMutex.Lock()
		defer listMutex.Unlock()
		key := commands[1]
		valuesToAdd := commands[2:]
		utils.ReverseSlice(valuesToAdd)
		list := []string{}
		list = append(list, valuesToAdd...)
		if val, ok := m[key]; !ok {
			m[key] = list
		} else {
			list = append(list, val...)
			m[key] = list
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
	}
	return ""
}

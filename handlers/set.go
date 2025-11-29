package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AYGA2K/dictago/types"
)

// It sets the value of a key, with optional expiration.
func Set(commands []string, m map[string]types.SetArg, replicas *types.ReplicaConns) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'set' command\r\n"
	}
	setArg := types.SetArg{
		Value:     commands[2],
		CreatedAt: time.Now(),
	}
	// Handle optional arguments like EX and PX for expiration.
	if len(commands) > 4 {
		switch strings.ToUpper(commands[3]) {
		case "EX":
			// EX specifies expiration in seconds.
			ex, err := strconv.Atoi(commands[4])
			if err != nil {
				return "-ERR value is not an integer or out of range\r\n"
			}
			setArg.Ex = ex
		case "PX":
			// PX specifies expiration in milliseconds.
			px, err := strconv.Atoi(commands[4])
			if err != nil {
				return "-ERR value is not an integer or out of range\r\n"
			}
			setArg.Px = px
		}
	}
	// Set the key to the new value.
	m[commands[1]] = setArg

	// Propagate the SET command to all replicas.
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

	// Return a simple string reply.
	return "+OK\r\n"
}

package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/AYGA2K/dictago/types"
)

// It sets the value of a key, with optional expiration.
func Set(commands []string, kvStore map[string]types.SetArg) string {
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
	kvStore[commands[1]] = setArg

	// Return a simple string reply.
	return "+OK\r\n"
}

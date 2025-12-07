package handlers

import (
	"fmt"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

func Get(commands []string, kvStore map[string]types.KVEntry) string {
	if len(commands) > 1 {
		// Check if the key exists in the map.
		if val, ok := kvStore[commands[1]]; ok {
			// Check if the key has expired.
			if !utils.IsExpired(val) {
				// If not expired, return the value as a bulk string.
				return fmt.Sprintf("$%d\r\n%s\r\n", len(val.Value), val.Value)
			}
		}
	}
	// If the key does not exist or has expired, return a null bulk string.
	return "$-1\r\n"
}

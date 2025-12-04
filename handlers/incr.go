package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AYGA2K/dictago/types"
)

// It increments the integer value of a key by one.
func Incr(commands []string, kvStore map[string]types.SetArg) string {
	if len(commands) < 2 {
		return "-ERR wrong number of arguments for 'incr' command\r\n"
	}
	key := commands[1]
	// If the key exists, increment its value.
	if val, ok := kvStore[key]; ok {
		intVal, err := strconv.Atoi(val.Value)
		if err != nil {
			return "-ERR value is not an integer or out of range\r\n"
		}
		intVal++
		val.Value = strconv.Itoa(intVal)
		kvStore[key] = val

		// Return the new value as an integer reply.
		return fmt.Sprintf(":%d\r\n", intVal)
	}
	// If the key does not exist, set it to 1.
	new := types.SetArg{
		Value:     "1",
		CreatedAt: time.Now(),
	}
	kvStore[key] = new

	return ":1\r\n"
}

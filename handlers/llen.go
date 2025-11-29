package handlers

import (
	"fmt"
	"sync"
)

// It returns the length of the list stored at a key.
func Llen(commands []string, m map[string][]string, listMutex *sync.Mutex) string {
	if len(commands) < 2 {
		return "-ERR wrong number of arguments for 'llen' command\r\n"
	}
	listMutex.Lock()
	defer listMutex.Unlock()
	key := commands[1]

	// If the key exists and holds a list, return its length.
	if val, ok := m[key]; ok {
		return fmt.Sprintf(":%d\r\n", len(val))
	} else {
		// If the key does not exist, it's treated as an empty list.
		return ":0\r\n"
	}
}

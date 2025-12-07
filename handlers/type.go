package handlers

import "github.com/AYGA2K/dictago/types"

// It returns the string representation of the type of the value stored at a key.
func Type(commands []string, kvStore map[string]types.KVEntry, listStore map[string][]string, streamStore map[string][]types.StreamEntry) string {
	if len(commands) < 2 {
		return "-ERR wrong number of arguments for 'type' command\r\n"
	}
	key := commands[1]
	// Check if the key exists in the string map.
	if _, ok := kvStore[key]; ok {
		return "+string\r\n"
	}
	// Check if the key exists in the list map.
	if _, ok := listStore[key]; ok {
		return "+list\r\n"
	}
	// Check if the key exists in the stream map.
	if _, ok := streamStore[key]; ok {
		return "+stream\r\n"
	}

	// If the key does not exist, return "none".
	return "+none\r\n"
}

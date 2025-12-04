package handlers

import (
	"fmt"
	"strconv"
	"sync"
)

// It returns the specified elements of the list stored at a key.
func Lrange(commands []string, listStore map[string][]string, listMutex *sync.Mutex) string {
	listMutex.Lock()
	defer listMutex.Unlock()
	if len(commands) < 4 {
		return "-ERR wrong number of arguments for 'lrange' command\r\n"
	}

	key := commands[1]
	val, ok := listStore[key]
	// If the key does not exist, return an empty array.
	if !ok {
		return "*0\r\n"
	}

	// Parse the start and end indices.
	start, err := strconv.Atoi(commands[2])
	if err != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}
	end, err := strconv.Atoi(commands[3])
	if err != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	n := len(val)

	// Handle negative indices. A negative index means start from the end of the list.
	if start < 0 {
		start = n + start
	}
	if end < 0 {
		end = n + end
	}

	// Clamp indices to the valid range of the list.
	if start < 0 {
		start = 0
	}
	if end >= n {
		end = n - 1
	}

	// If the range is invalid or empty, return an empty array.
	if start > end || start >= n {
		return "*0\r\n"
	}

	// Build the response with the elements in the specified range.
	response := fmt.Sprintf("*%d\r\n", end-start+1)
	for _, v := range val[start : end+1] {
		response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}
	return response
}

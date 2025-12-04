package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/utils"
)

// It prepends one or multiple values to a list.
func Lpush(commands []string, listStore map[string][]string, listMutex *sync.Mutex) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'lpush' command\r\n"
	}
	listMutex.Lock()
	defer listMutex.Unlock()
	key := commands[1]
	valuesToAdd := commands[2:]
	// Reverse the slice of values to add, so they are prepended in the correct order.
	utils.ReverseSlice(valuesToAdd)
	list := []string{}
	list = append(list, valuesToAdd...)
	// If the key does not exist, create a new list.
	if val, ok := listStore[key]; !ok {
		listStore[key] = list
	} else {
		// If the key exists, prepend the new values to the existing list.
		list = append(list, val...)
		listStore[key] = list
	}

	// Return the length of the list after the push operation.
	return fmt.Sprintf(":%d\r\n", len(list))
}

package handlers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// It's a blocking pop from the head of a list.
func Blpop(commands []string, listStore map[string][]string, listMutex *sync.Mutex, blockingClients map[string][]chan any) string {
	if len(commands) < 3 {
		return "wrong number of arguments for 'blpop' command\r\n"
	}

	key := commands[1]
	timeout, err := strconv.ParseFloat(commands[2], 64)
	if err != nil {
		return "timeout is not a float or out of range\r\n"
	}

	listMutex.Lock()

	// If the list exists and is not empty, pop the first element and return.
	if val, ok := listStore[key]; ok && len(val) > 0 {
		popped := val[0]
		listStore[key] = val[1:]
		listMutex.Unlock()
		return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(popped), popped)
	}

	// If we are here, the list is empty, so we need to block.
	// Create a channel to wait on.
	waitChan := make(chan any)
	blockingClients[key] = append(blockingClients[key], waitChan)

	listMutex.Unlock()

	// Set up a context for timeout.
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Second)))
	} else {
		// A timeout of 0 means block indefinitely.
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Block until either the channel is signaled or the timeout is reached.
	select {
	case <-waitChan:
		// We were woken up by a pusher (e.g., RPUSH or LPUSH).
		listMutex.Lock()
		defer listMutex.Unlock()
		// The list should now have an element. Pop it and return.
		if val, ok := listStore[key]; ok && len(val) > 0 {
			popped := val[0]
			listStore[key] = val[1:]
			return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(popped), popped)
		}
	case <-ctx.Done():
		// Timeout occurred.
		listMutex.Lock()
		// Remove our channel from the waiting list to avoid memory leaks.
		clients, ok := blockingClients[key]
		if ok {
			for i, c := range clients {
				if c == waitChan {
					blockingClients[key] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
		}
		listMutex.Unlock()
		// Return a null bulk string reply as per Redis protocol.
		return "*-1\r\n"
	}
	// Should not be reached.
	return ""
}

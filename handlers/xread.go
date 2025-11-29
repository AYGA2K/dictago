package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

// It reads from one or more streams, optionally blocking if no new data is available.
func Xread(commands []string, streams map[string][]types.StreamEntry, streamsMutex *sync.Mutex, waitingClients map[string][]chan any) string {
	// Parse the arguments from the command.
	block, timeout, keys, ids, err := parseXreadArgs(commands)
	if err != nil {
		return fmt.Sprintf("-ERR %v\r\n", err)
	}
	// If the special ID "$" is used, it means we should start reading from the last entry in the stream.
	if block && len(ids) == 1 && ids[0] == "$" {
		if val, ok := streams[keys[0]]; ok {
			ids[0] = val[len(val)-1].ID
		} else {
			// If the stream doesn't exist, we can't get the last ID, so return nil.
			return "*-1\r\n"
		}
	}

	// Process the streams and get the initial response.
	streamsMutex.Lock()
	response, hasData := processXread(keys, ids, streams)
	streamsMutex.Unlock()

	// If there is data or if it's not a blocking call, return the response immediately.
	if hasData || !block {
		return response
	}

	// No data, and it's a blocking call. We need to wait.
	waitChan := make(chan any)
	for _, key := range keys {
		waitingClients[key] = append(waitingClients[key], waitChan)
	}

	// Set up a context for the timeout.
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout*float64(time.Millisecond)))
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Block until either we are notified of new data or the timeout occurs.
	select {
	case <-waitChan:
		// We were woken up by an XADD. Reprocess the streams.
		streamsMutex.Lock()
		response, _ := processXread(keys, ids, streams)
		streamsMutex.Unlock()
		return response
	case <-ctx.Done():
		// Timeout occurred.
		streamsMutex.Lock()
		// Remove our channel from the waiting lists to avoid memory leaks.
		for _, key := range keys {
			clients, ok := waitingClients[key]
			if ok {
				for i, c := range clients {
					if c == waitChan {
						waitingClients[key] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
			}
		}
		streamsMutex.Unlock()
		return "*-1\r\n"
	}
}

// parseXreadArgs parses the arguments for the XREAD command.
func parseXreadArgs(commands []string) (block bool, timeout float64, keys []string, ids []string, err error) {
	if len(commands) < 4 {
		return false, 0, nil, nil, fmt.Errorf("wrong number of arguments for 'xread' command")
	}

	i := 1
	// Check for the BLOCK option.
	if strings.ToUpper(commands[i]) == "BLOCK" {
		block = true
		i++
		timeout, err = strconv.ParseFloat(commands[i], 64)
		if err != nil {
			return false, 0, nil, nil, fmt.Errorf("timeout is not a float or out of range")
		}
		i++
	}

	// The STREAMS keyword must be present.
	if strings.ToUpper(commands[i]) != "STREAMS" {
		return false, 0, nil, nil, fmt.Errorf("expected STREAMS keyword")
	}
	i++

	// The remaining arguments are the keys and their corresponding IDs.
	numKeys := (len(commands) - i) / 2
	keys = commands[i : i+numKeys]
	ids = commands[i+numKeys:]

	if len(keys) != len(ids) {
		return false, 0, nil, nil, fmt.Errorf("uneven number of streams and IDs")
	}

	return block, timeout, keys, ids, nil
}

// processXread processes the streams and generates the response for XREAD.
func processXread(keys []string, ids []string, streams map[string][]types.StreamEntry) (string, bool) {
	var response strings.Builder
	hasData := false

	// The response is an array of streams.
	response.WriteString(fmt.Sprintf("*%d\r\n", len(keys)))

	for i := range keys {
		key := keys[i]
		startID := ids[i]

		stream, ok := streams[key]
		// If the stream does not exist, return an empty array for that stream.
		if !ok {
			response.WriteString("*2\r\n")
			response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(key), key))
			response.WriteString("*0\r\n")
			continue
		}

		// Find all entries with an ID greater than the start ID.
		var results []types.StreamEntry
		for _, entry := range stream {
			isGreater, err := utils.IsGreaterThan(entry.ID, startID)
			if err != nil {
				return fmt.Sprintf("-ERR %v\r\n", err), false
			}
			if isGreater {
				results = append(results, entry)
			}
		}

		if len(results) > 0 {
			hasData = true
		}

		// Format the stream and its entries.
		response.WriteString("*2\r\n")
		response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(key), key))
		response.WriteString(fmt.Sprintf("*%d\r\n", len(results)))

		for _, entry := range results {
			response.WriteString("*2\r\n")
			response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(entry.ID), entry.ID))
			response.WriteString(fmt.Sprintf("*%d\r\n", len(entry.Data)*2))
			for field, value := range entry.Data {
				response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(field), field))
				response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
			}
		}
	}

	return response.String(), hasData
}

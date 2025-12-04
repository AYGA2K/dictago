package handlers

import (
	"fmt"
	"sync"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
)

// It returns a range of entries from a stream.
func Xrange(commands []string, streamStore map[string][]types.StreamEntry, streamsMutex *sync.Mutex) string {
	streamsMutex.Lock()
	defer streamsMutex.Unlock()

	if len(commands) < 4 {
		return "-ERR wrong number of arguments for 'xrange' command\r\n"
	}

	key := commands[1]
	startID := commands[2]
	endID := commands[3]

	// Parse the start and end IDs.
	startMs, startSeq, err := utils.ParseStreamID(startID)
	if err != nil {
		return fmt.Sprintf("-ERR Invalid start ID: %v\r\n", err)
	}

	endMs, endSeq, err := utils.ParseStreamID(endID)
	if err != nil {
		return fmt.Sprintf("-ERR Invalid end ID: %v\r\n", err)
	}

	// If the stream does not exist, return an empty array.
	stream, ok := streamStore[key]
	if !ok {
		return "*0\r\n"
	}

	// Iterate through the stream and collect entries that fall within the specified range.
	var results []types.StreamEntry
	for _, entry := range stream {
		if utils.InRange(entry, startMs, startSeq, endMs, endSeq) {
			results = append(results, entry)
		}
	}

	// Format the results as a RESP array.
	resp := fmt.Sprintf("*%d\r\n", len(results))
	for _, entry := range results {
		resp += "*2\r\n"
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(entry.ID), entry.ID)
		resp += fmt.Sprintf("*%d\r\n", len(entry.Data)*2)
		for k, v := range entry.Data {
			resp += fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n", len(k), k, len(v), v)
		}
	}
	return resp
}

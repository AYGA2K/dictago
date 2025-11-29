package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseRespArray parses a RESP (Redis Serialization Protocol) array from a bufio.Reader.
// It returns the parsed commands as a slice of strings, the total bytes read, and any error encountered.
func ParseRespArray(reader *bufio.Reader) ([]string, int, error) {
	bytesRead := 0
	commands := []string{}

	// Read the first line, which should be the array header.
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, 0, err
	}
	bytesRead += len(header)

	// A RESP array starts with '*'.
	if header[0] == '*' {
		// Get the number of elements in the array.
		elemsNum, err := strconv.Atoi(strings.TrimSpace(header[1:]))
		if err != nil {
			return nil, bytesRead, err
		}
		// Parse each element in the array.
		for range elemsNum {
			// Read the bulk string length line.
			lenLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, bytesRead, err
			}
			bytesRead += len(lenLine)
			// A bulk string length line starts with '$'.
			if lenLine[0] != '$' {
				return nil, bytesRead, fmt.Errorf("invalid bulk string length line")
			}

			// Get the length of the bulk string.
			length, err := strconv.Atoi(strings.TrimSpace(lenLine[1:]))
			if err != nil {
				return nil, bytesRead, err
			}

			// Read the bulk string itself.
			bytes := make([]byte, length)
			n, err := io.ReadFull(reader, bytes)
			if err != nil {
				return nil, bytesRead, err
			}
			bytesRead += n

			commands = append(commands, string(bytes))

			// Discard the trailing \r\n.
			discarded, err := reader.Discard(2)
			if err != nil {
				return nil, bytesRead, err
			}
			bytesRead += discarded
		}
	} else {
		return nil, bytesRead, fmt.Errorf("unsupported command format: not a RESP array")
	}
	return commands, bytesRead, nil
}

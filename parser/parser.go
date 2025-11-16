package parser

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func ParseRespArray(reader *bufio.Reader) ([]string, error) {
	commands := []string{}
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if header[0] == '*' {
		elemsNum, err := strconv.Atoi(strings.TrimSpace(header[1:]))
		if err != nil {
			return nil, err
		}
		for range elemsNum {
			lenLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			if lenLine[0] != '$' {
				return nil, fmt.Errorf("invalid bulk string length line")
			}

			length, err := strconv.Atoi(strings.TrimSpace(lenLine[1:]))
			if err != nil {
				return nil, err
			}

			bytes := make([]byte, length)
			_, err = reader.Read(bytes)
			if err != nil {
				return nil, err
			}

			commands = append(commands, string(bytes))

			// Discard \r\n
			_, err = reader.Discard(2)
			if err != nil {
				return nil, err
			}
		}
	}
	return commands, err
}

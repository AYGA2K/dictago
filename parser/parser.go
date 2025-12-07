package parser

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/AYGA2K/dictago/types"
	"github.com/AYGA2K/dictago/utils"
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

func ParseRDB(path string) (*types.ParsedRDB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	p := &types.ParsedRDB{}

	// --- 1. HEADER ("REDIS0011") ---
	header := make([]byte, 9)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, err
	}
	if string(header)[:5] != "REDIS" {
		return nil, errors.New("invalid RDB magic header")
	}

	// --- 2. METADATA SECTION (FA ...) ---
	for {
		b, err := utils.ReadByte(f)
		if err != nil {
			return nil, err
		}

		if b == 0xFA { // metadata start
			_, err := utils.ReadString(f) // metadata name (ignored)
			if err != nil {
				return nil, err
			}
			_, err = utils.ReadString(f) // metadata value (ignored)
			if err != nil {
				return nil, err
			}
			continue
		}

		// If it's not FA, it's part of DB section — go back 1 byte
		if _, err := f.Seek(-1, io.SeekCurrent); err != nil {
			return nil, err
		}
		break
	}

	// --- 3. DATABASE SECTION ---
	for {
		marker, err := utils.ReadByte(f)
		if err != nil {
			return nil, err
		}

		switch marker {

		case 0xFE: // database start
			_, _, err := utils.ReadLength(f) // database index (ignored)
			if err != nil {
				return nil, err
			}

		case 0xFB: // hash table sizes
			_, _, _ = utils.ReadLength(f) // key-value table size
			_, _, _ = utils.ReadLength(f) // expires table size

		case 0xFC, 0xFD, 0x00:
			// The marker is actually part of a key record
			if _, err := f.Seek(-1, io.SeekCurrent); err != nil {
				return nil, err
			}
			key, err := parseKey(f)
			if err != nil {
				return nil, err
			}
			p.Keys = append(p.Keys, key)

		case 0xFF:
			// END OF FILE SECTION
			checksum := make([]byte, 8)
			if _, err := io.ReadFull(f, checksum); err != nil {
				return nil, err
			}

			return p, nil

		default:
			return nil, fmt.Errorf("unexpected marker: %x", marker)
		}
	}

}

func parseKey(r io.Reader) (types.RDBKey, error) {
	key := types.RDBKey{}
	var err error

	// OPTIONAL EXPIRE
	first, err := utils.ReadByte(r)
	if err != nil {
		return key, err
	}

	switch first {
	case 0xFC: // expire in ms
		buf := make([]byte, 8)
		if _, err := io.ReadFull(r, buf); err != nil {
			return key, err
		}
		key.ExpireAt = int64(binary.LittleEndian.Uint64(buf))
		key.HasExpire = true
		key.Type = "ms"

		first, err = utils.ReadByte(r) // next byte is type
		if err != nil {
			return key, err
		}

	case 0xFD: // expire in seconds
		buf := make([]byte, 4)
		if _, err := io.ReadFull(r, buf); err != nil {
			return key, err
		}
		secs := binary.LittleEndian.Uint32(buf)
		key.ExpireAt = int64(secs)
		key.HasExpire = true
		key.Type = "s"

		first, err = utils.ReadByte(r) // next byte is type
		if err != nil {
			return key, err
		}
	}

	// VALUE TYPE (only support "string" = 0x00)
	if first != 0x00 {
		return key, fmt.Errorf("unsupported value type: %x", first)
	}

	// KEY
	key.Key, err = utils.ReadString(r)
	if err != nil {
		return key, err
	}

	// VALUE
	key.Value, err = utils.ReadString(r)
	if err != nil {
		return key, err
	}

	return key, nil
}

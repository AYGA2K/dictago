package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// It is used by replicas to acknowledge the amount of data they have processed.
func Replconf(con net.Conn, commands []string, replicaOffset int) string {
	// If the command is REPLCONF GETACK *, the master is asking the replica
	// to acknowledge the amount of data it has processed.
	if len(commands) == 3 && strings.ToUpper(commands[1]) == "GETACK" && strings.ToUpper(commands[2]) == "*" {
		// The replica responds with its replication offset.
		strOffset := strconv.Itoa(replicaOffset)
		return fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%d\r\n", len(strOffset), replicaOffset)
	}
	// For other REPLCONF commands, just respond with OK.
	return "+OK\r\n"
}

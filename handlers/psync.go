package handlers

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/AYGA2K/dictago/types"
)

// This is part of the replication process, allowing a replica to synchronize with the master.
func Psync(con net.Conn, masterReplid string, replicas *types.ReplicaConns) {
	// Respond with a FULLRESYNC message, indicating a full synchronization will occur.
	response := fmt.Sprintf("+FULLRESYNC %s 0\r\n", masterReplid)
	con.Write([]byte(response))

	// Read the RDB file content.
	// NOTE: This assumes an empty RDB file named "rdb" exists in the root directory.
	hexContent, err := os.ReadFile("rdb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading RDB file: %v\n", err)
		return
	}

	// Decode the hex-encoded RDB file content.
	decodedContent, err := hex.DecodeString(string(hexContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding RDB file: %v\n", err)
		return
	}

	// Send the RDB file to the replica as a bulk string.
	fmt.Fprintf(con, "$%d\r\n%s", len(decodedContent), decodedContent)

	// Add the connection to the list of replicas.
	replicas.Lock()
	replicas.Conns = append(replicas.Conns, con)
	replicas.Unlock()
}

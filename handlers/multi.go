package handlers

import "github.com/AYGA2K/dictago/types"

// It marks the start of a transaction block.
func Multi(client *types.Client) string {
	// Set the InMulti flag on the client to true.
	client.InMulti = true
	// Return a simple string reply.
	return "+OK\r\n"
}

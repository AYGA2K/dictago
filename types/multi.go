package types

// Client represents a connected client and its state.
type Client struct {
	// InMulti is true if the client is in a MULTI...EXEC block.
	InMulti bool
	// Commands is a queue of commands to be executed in a transaction.
	Commands [][]string
}

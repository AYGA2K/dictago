package types

import (
	"net"
	"sync"
	"time"
)

type KVEntry struct {
	Value string
	// Ex is the expiration time in seconds.
	Ex int
	// Px is the expiration time in milliseconds.
	Px        int
	CreatedAt time.Time
}

type StreamEntry struct {
	ID   string
	Data map[string]string
}

type ParsedRDB struct {
	Keys []RDBKey
}

type RDBKey struct {
	Key       string
	Value     string
	ExpireAt  int64
	HasExpire bool
	Type      string
}

type ReplicaConns struct {
	sync.Mutex
	Conns []net.Conn
}

// Client represents a connected client and its state.
type Client struct {
	// InMulti is true if the client is in a MULTI...EXEC block.
	InMulti bool
	// Commands is a queue of commands to be executed in a transaction.
	Commands  [][]string
	Connected bool
}

type User struct {
	Name      string
	Flags     []string
	Passwords []string
}

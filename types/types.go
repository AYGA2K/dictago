package types

import (
	"time"
)

type SetArg struct {
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

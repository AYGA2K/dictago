package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AYGA2K/dictago/types"
)

// ParseStreamID parses a stream ID string into its milliseconds and sequence number parts.
// It also handles the special IDs "+" (infinity) and "-" (zero).
func ParseStreamID(id string) (int64, int64, error) {
	if id == "+" {
		return 9223372036854775807, 9223372036854775807, nil
	}
	if id == "-" {
		return 0, 0, nil
	}
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid stream ID format: %s", id)
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid milliseconds part in stream ID: %s", parts[0])
	}
	seq, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sequence part in stream ID: %s", parts[1])
	}
	return ms, seq, nil
}

// InRange checks if a stream entry's ID is within a given range (inclusive).
func InRange(entry types.StreamEntry, startMs, startSeq, endMs, endSeq int64) bool {
	entryMs, entrySeq, err := ParseStreamID(entry.ID)
	if err != nil {
		return false
	}

	afterStart := entryMs > startMs || (entryMs == startMs && entrySeq >= startSeq)

	beforeEnd := entryMs < endMs || (entryMs == endMs && entrySeq <= endSeq)

	return afterStart && beforeEnd
}

// IsGreaterThan checks if an entry ID is greater than a start ID.
func IsGreaterThan(entryID, startID string) (bool, error) {
	entryMs, entrySeq, err := ParseStreamID(entryID)
	if err != nil {
		return false, err
	}
	startMs, startSeq, err := ParseStreamID(startID)
	if err != nil {
		return false, err
	}

	if entryMs > startMs {
		return true, nil
	}
	if entryMs == startMs && entrySeq > startSeq {
		return true, nil
	}
	return false, nil
}

package main_test

import (
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestStreamCommands(t *testing.T) {
	streams := make(map[string][]types.StreamEntry)
	streamsMutex := &sync.Mutex{}
	replicas := &types.ReplicaConns{}
	waitingClients := make(map[string][]chan any)

	// Test XADD
	xaddCmd := []string{"XADD", "mystream", "0-1", "name", "John"}
	xaddResult := handlers.Xadd(xaddCmd, streams, streamsMutex, waitingClients, replicas)
	expectedXaddResult := "$3\r\n0-1\r\n"
	if xaddResult != expectedXaddResult {
		t.Errorf("XADD: Expected %q, got %q", expectedXaddResult, xaddResult)
	}

	// Test XADD with auto-generated ID
	xaddCmd2 := []string{"XADD", "mystream", "*", "surname", "Doe"}
	xaddResult2 := handlers.Xadd(xaddCmd2, streams, streamsMutex, waitingClients, replicas)
	if len(xaddResult2) < 1 || xaddResult2[0] != '$' {
		t.Errorf("XADD with auto-generated ID: Expected a bulk string, got %q", xaddResult2)
	}

	// Test XRANGE
	xrangeCmd := []string{"XRANGE", "mystream", "-", "+"}
	xrangeResult := handlers.Xrange(xrangeCmd, streams, streamsMutex)
	expectedXrangeResult := "*2\r\n*2\r\n$3\r\n0-1\r\n*2\r\n$4\r\nname\r\n$4\r\nJohn\r\n*2\r\n"
	if len(xrangeResult) < len(expectedXrangeResult) {
		t.Errorf("XRANGE: Expected a result with at least %d characters, got %d", len(expectedXrangeResult), len(xrangeResult))
	}

	// Test XREAD
	xreadCmd := []string{"XREAD", "streams", "mystream", "0-0"}
	xreadResult := handlers.Xread(xreadCmd, streams, streamsMutex, waitingClients)
	expectedXreadResult := "*1\r\n*2\r\n$8\r\nmystream\r\n*2\r\n*2\r\n$3\r\n0-1\r\n*2\r\n$4\r\nname\r\n$4\r\nJohn\r\n*2\r\n"
	if len(xreadResult) < len(expectedXreadResult) {
		t.Errorf("XREAD: Expected a result with at least %d characters, got %d", len(expectedXreadResult), len(xreadResult))
	}
}

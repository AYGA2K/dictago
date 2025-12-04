package main_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestStreamCommands(t *testing.T) {
	streams := make(map[string][]types.StreamEntry)
	streamsMutex := &sync.Mutex{}
	waitingClients := make(map[string][]chan any)

	// Test XADD
	xaddCmd := []string{"XADD", "mystream", "0-1", "name", "John"}
	xaddResult := handlers.Xadd(xaddCmd, streams, streamsMutex, waitingClients)
	expectedXaddResult := "$3\r\n0-1\r\n"
	if xaddResult != expectedXaddResult {
		t.Errorf("XADD: Expected %q, got %q", expectedXaddResult, xaddResult)
	}

	// Test XADD with auto-generated ID
	xaddCmd2 := []string{"XADD", "mystream", "*", "surname", "Doe"}
	xaddResult2 := handlers.Xadd(xaddCmd2, streams, streamsMutex, waitingClients)
	if len(xaddResult2) < 1 || xaddResult2[0] != '$' {
		t.Errorf("XADD with auto-generated ID: Expected a bulk string, got %q", xaddResult2)
	}

	// Test XRANGE
	xrangeCmd := []string{"XRANGE", "mystream", "-", "+"}
	xrangeResult := handlers.Xrange(xrangeCmd, streams, streamsMutex)
	if !strings.HasPrefix(xrangeResult, "*2\r\n") {
		t.Errorf("XRANGE: Expected an array of 2 elements, got %q", xrangeResult)
	}
	if !strings.Contains(xrangeResult, "$4\r\nname\r\n$4\r\nJohn\r\n") {
		t.Errorf("XRANGE: Expected to find 'name' and 'John' in the response, got %q", xrangeResult)
	}
	if !strings.Contains(xrangeResult, "$7\r\nsurname\r\n$3\r\nDoe\r\n") {
		t.Errorf("XRANGE: Expected to find 'surname' and 'Doe' in the response, got %q", xrangeResult)
	}

	// Test XREAD
	xreadCmd := []string{"XREAD", "streams", "mystream", "0-0"}
	xreadResult := handlers.Xread(xreadCmd, streams, streamsMutex, waitingClients)
	if !strings.HasPrefix(xreadResult, "*1\r\n*2\r\n$8\r\nmystream\r\n") {
		t.Errorf("XREAD: Expected a response for one stream, got %q", xreadResult)
	}
	if !strings.Contains(xreadResult, "$4\r\nname\r\n$4\r\nJohn\r\n") {
		t.Errorf("XREAD: Expected to find 'name' and 'John' in the response, got %q", xreadResult)
	}
	if !strings.Contains(xreadResult, "$7\r\nsurname\r\n$3\r\nDoe\r\n") {
		t.Errorf("XREAD: Expected to find 'surname' and 'Doe' in the response, got %q", xreadResult)
	}
}

package main_test

import (
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
)

func TestListCommands(t *testing.T) {
	mlist := make(map[string][]string)
	listMutex := &sync.Mutex{}
	waitingClients := make(map[string][]chan any)

	// Test LPUSH
	lpushCmd := []string{"LPUSH", "mylist", "world"}
	lpushResult := handlers.Lpush(lpushCmd, mlist, listMutex)
	expectedLpushResult := ":1\r\n"
	if lpushResult != expectedLpushResult {
		t.Errorf("LPUSH: Expected %q, got %q", expectedLpushResult, lpushResult)
	}

	lpushCmd2 := []string{"LPUSH", "mylist", "hello"}
	handlers.Lpush(lpushCmd2, mlist, listMutex)

	// Test RPUSH
	rpushCmd := []string{"RPUSH", "mylist", "!"}
	rpushResult := handlers.Rpush(rpushCmd, mlist, listMutex, waitingClients)
	expectedRpushResult := ":3\r\n"
	if rpushResult != expectedRpushResult {
		t.Errorf("RPUSH: Expected %q, got %q", expectedRpushResult, rpushResult)
	}

	// Test LLEN
	llenCmd := []string{"LLEN", "mylist"}
	llenResult := handlers.Llen(llenCmd, mlist, listMutex)
	expectedLlenResult := ":3\r\n"
	if llenResult != expectedLlenResult {
		t.Errorf("LLEN: Expected %q, got %q", expectedLlenResult, llenResult)
	}

	// Test LRANGE
	lrangeCmd := []string{"LRANGE", "mylist", "0", "-1"}
	lrangeResult := handlers.Lrange(lrangeCmd, mlist, listMutex)
	expectedLrangeResult := "*3\r\n$5\r\nhello\r\n$5\r\nworld\r\n$1\r\n!\r\n"
	if lrangeResult != expectedLrangeResult {
		t.Errorf("LRANGE: Expected %q, got %q", expectedLrangeResult, lrangeResult)
	}

	// Test LPOP
	lpopCmd := []string{"LPOP", "mylist"}
	lpopResult := handlers.Lpop(lpopCmd, mlist, listMutex)
	expectedLpopResult := "$5\r\nhello\r\n"
	if lpopResult != expectedLpopResult {
		t.Errorf("LPOP: Expected %q, got %q", expectedLpopResult, lpopResult)
	}

	// Test LPOP on empty list
	handlers.Lpop(lpopCmd, mlist, listMutex)
	handlers.Lpop(lpopCmd, mlist, listMutex)
	lpopResultEmpty := handlers.Lpop(lpopCmd, mlist, listMutex)
	expectedLpopResultEmpty := "$-1\r\n"
	if lpopResultEmpty != expectedLpopResultEmpty {
		t.Errorf("LPOP on empty list: Expected %q, got %q", expectedLpopResultEmpty, lpopResultEmpty)
	}
}

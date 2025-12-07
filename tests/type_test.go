package main_test

import (
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestTypeCommand(t *testing.T) {
	m := make(map[string]types.KVEntry)
	mlist := make(map[string][]string)
	streams := make(map[string][]types.StreamEntry)
	listMutex := &sync.Mutex{}

	// Test TYPE on string
	setCmd := []string{"SET", "mykey", "myvalue"}
	handlers.Set(setCmd, m)
	typeCmd1 := []string{"TYPE", "mykey"}
	typeResult1 := handlers.Type(typeCmd1, m, mlist, streams)
	expectedTypeResult1 := "+string\r\n"
	if typeResult1 != expectedTypeResult1 {
		t.Errorf("TYPE string: Expected %q, got %q", expectedTypeResult1, typeResult1)
	}

	// Test TYPE on list
	lpushCmd := []string{"LPUSH", "mylist", "world"}
	handlers.Lpush(lpushCmd, mlist, listMutex)
	typeCmd2 := []string{"TYPE", "mylist"}
	typeResult2 := handlers.Type(typeCmd2, m, mlist, streams)
	expectedTypeResult2 := "+list\r\n"
	if typeResult2 != expectedTypeResult2 {
		t.Errorf("TYPE list: Expected %q, got %q", expectedTypeResult2, typeResult2)
	}

	// Test TYPE on stream
	xaddCmd := []string{"XADD", "mystream", "0-1", "name", "John"}
	streamsMutex := &sync.Mutex{}
	waitingClients := make(map[string][]chan any)
	handlers.Xadd(xaddCmd, streams, streamsMutex, waitingClients)
	typeCmd3 := []string{"TYPE", "mystream"}
	typeResult3 := handlers.Type(typeCmd3, m, mlist, streams)
	expectedTypeResult3 := "+stream\r\n"
	if typeResult3 != expectedTypeResult3 {
		t.Errorf("TYPE stream: Expected %q, got %q", expectedTypeResult3, typeResult3)
	}

	// Test TYPE on non-existent key
	typeCmd4 := []string{"TYPE", "nonexistentkey"}
	typeResult4 := handlers.Type(typeCmd4, m, mlist, streams)
	expectedTypeResult4 := "+none\r\n"
	if typeResult4 != expectedTypeResult4 {
		t.Errorf("TYPE none: Expected %q, got %q", expectedTypeResult4, typeResult4)
	}
}

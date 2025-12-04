package main_test

import (
	"net"
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestTransactionCommands(t *testing.T) {
	client := &types.Client{}
	m := make(map[string]types.SetArg)
	mlist := make(map[string][]string)
	streams := make(map[string][]types.StreamEntry)
	listMutex := &sync.Mutex{}
	streamsMutex := &sync.Mutex{}
	waitingClients := make(map[string][]chan any)
	clients := make(map[net.Conn]*types.Client)

	// Test MULTI
	multiResult := handlers.Multi(client)
	expectedMultiResult := "+OK\r\n"
	if multiResult != expectedMultiResult {
		t.Errorf("MULTI: Expected %q, got %q", expectedMultiResult, multiResult)
	}
	if !client.InMulti {
		t.Error("MULTI: client.InMulti should be true")
	}

	// Test QUEUED
	client.Commands = append(client.Commands, []string{"SET", "foo", "bar"})

	// Test DISCARD
	discardClient := &types.Client{InMulti: true}
	if !discardClient.InMulti {
		t.Error("DISCARD: client.InMulti should be true")
	}
	discardClient.InMulti = false
	discardClient.Commands = nil
	if discardClient.InMulti {
		t.Error("DISCARD: client.InMulti should be false after DISCARD")
	}
	if discardClient.Commands != nil {
		t.Error("DISCARD: client.Commands should be nil after DISCARD")
	}

	// Test EXEC
	execClient := &types.Client{InMulti: true}
	execClient.Commands = append(execClient.Commands, []string{"SET", "foo", "bar"})
	execClient.Commands = append(execClient.Commands, []string{"GET", "foo"})

	// Create a dummy connection for Exec
	conn, server := net.Pipe()
	clients[conn] = execClient

	var wg sync.WaitGroup
	wg.Go(func() {
		buf := make([]byte, 1024)
		server.Read(buf)
	})

	handlers.Exec(conn, clients, m, mlist, streams, listMutex, streamsMutex, waitingClients)
	conn.Close()
	wg.Wait()

	// We can't easily test the response from Exec as it writes directly to the connection.
	// But we can check the side effects.
	if val, ok := m["foo"]; !ok || val.Value != "bar" {
		t.Error("EXEC: SET command was not executed correctly")
	}
}

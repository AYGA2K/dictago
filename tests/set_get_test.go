package main_test

import (
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestSetAndGet(t *testing.T) {
	m := make(map[string]types.SetArg)
	replicas := &types.ReplicaConns{}

	// Test SET
	setCmd := []string{"SET", "mykey", "myvalue"}
	setResult := handlers.Set(setCmd, m, replicas)
	expectedSetResult := "+OK\r\n"
	if setResult != expectedSetResult {
		t.Errorf("Expected %q, got %q", expectedSetResult, setResult)
	}

	// Test GET
	getCmd := []string{"GET", "mykey"}
	getResult := handlers.Get(getCmd, m)
	expectedGetResult := "$7\r\nmyvalue\r\n"
	if getResult != expectedGetResult {
		t.Errorf("Expected %q, got %q", expectedGetResult, getResult)
	}

	// Test GET non-existent key
	getCmdNonExistent := []string{"GET", "nonexistentkey"}
	getResultNonExistent := handlers.Get(getCmdNonExistent, m)
	expectedGetResultNonExistent := "$-1\r\n"
	if getResultNonExistent != expectedGetResultNonExistent {
		t.Errorf("Expected %q, got %q", expectedGetResultNonExistent, getResultNonExistent)
	}
}

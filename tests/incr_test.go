package main_test

import (
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestIncr(t *testing.T) {
	m := make(map[string]types.SetArg)
	replicas := &types.ReplicaConns{}

	// Test INCR on non-existent key
	incrCmd1 := []string{"INCR", "mykey"}
	incrResult1 := handlers.Incr(incrCmd1, m, replicas)
	expectedIncrResult1 := ":1\r\n"
	if incrResult1 != expectedIncrResult1 {
		t.Errorf("Expected %q, got %q", expectedIncrResult1, incrResult1)
	}

	// Test INCR on existing key
	incrCmd2 := []string{"INCR", "mykey"}
	incrResult2 := handlers.Incr(incrCmd2, m, replicas)
	expectedIncrResult2 := ":2\r\n"
	if incrResult2 != expectedIncrResult2 {
		t.Errorf("Expected %q, got %q", expectedIncrResult2, incrResult2)
	}

	// Test INCR on a key with a non-integer value
	setCmd := []string{"SET", "stringkey", "hello"}
	handlers.Set(setCmd, m, replicas)
	incrCmd3 := []string{"INCR", "stringkey"}
	incrResult3 := handlers.Incr(incrCmd3, m, replicas)
	expectedIncrResult3 := "-ERR value is not an integer or out of range\r\n"
	if incrResult3 != expectedIncrResult3 {
		t.Errorf("Expected %q, got %q", expectedIncrResult3, incrResult3)
	}
}

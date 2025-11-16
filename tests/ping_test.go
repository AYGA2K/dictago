package main_test

import (
	"testing"

	"github.com/AYGA2K/dictago/handlers"
)

func TestPing(t *testing.T) {
	result := handlers.Ping()
	expected := "+PONG\r\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

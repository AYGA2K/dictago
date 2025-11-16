package main_test

import (
	"testing"

	"github.com/AYGA2K/dictago/handlers"
)

func TestEcho(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "Simple echo",
			input:    []string{"ECHO", "hello"},
			expected: "$5\r\nhello\r\n",
		},
		{
			name:     "Echo with multiple spaces",
			input:    []string{"ECHO", "hello world"},
			expected: "$11\r\nhello world\r\n",
		},
		{
			name:     "Empty string",
			input:    []string{"ECHO", ""},
			expected: "$0\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.Echo(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestAuth(t *testing.T) {
	// Setup a default user with a password
	password := "mypassword"
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	defaultUserWithPass := types.User{
		Name:      "default",
		Flags:     []string{"on"},
		Passwords: []string{passwordHash},
	}

	// Setup a default user with nopass flag
	defaultUserNoPass := types.User{
		Name:      "default",
		Flags:     []string{"on", "nopass"},
		Passwords: []string{},
	}

	tests := []struct {
		name     string
		commands []string
		user     types.User
		expected string
	}{
		{
			name:     "Successful authentication with correct password",
			commands: []string{"AUTH", "default", password},
			user:     defaultUserWithPass,
			expected: "+OK\r\n",
		},
		{
			name:     "Failed authentication with wrong password",
			commands: []string{"AUTH", "default", "wrongpassword"},
			user:     defaultUserWithPass,
			expected: "-WRONGPASS invalid username-password pair or user is disabled.\r\n",
		},
		{
			name:     "Successful authentication for nopass user",
			commands: []string{"AUTH", "default", ""}, // Password can be empty for nopass
			user:     defaultUserNoPass,
			expected: "+OK\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.Auth(tt.commands, tt.user)
			if result != tt.expected {
				t.Errorf("For test case %s: Expected %q, got %q", tt.name, tt.expected, result)
			}
		})
	}
}

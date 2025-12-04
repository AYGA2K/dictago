package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"

	"github.com/AYGA2K/dictago/handlers"
	"github.com/AYGA2K/dictago/types"
)

func TestAcl(t *testing.T) {

	userMutex := &sync.Mutex{}

	tests := []struct {
		name         string
		commands     []string
		initialUser  *types.User
		expected     string
		expectedUser *types.User
	}{
		{
			name:     "ACL WHOAMI",
			commands: []string{"ACL", "WHOAMI"},
			initialUser: &types.User{
				Name:      "default",
				Flags:     []string{"on"},
				Passwords: []string{},
			},
			expected: "$7\r\ndefault\r\n",
		},
		{
			name:     "ACL GETUSER default",
			commands: []string{"ACL", "GETUSER", "default"},
			initialUser: &types.User{
				Name:      "default",
				Flags:     []string{"on"},
				Passwords: []string{},
			},
			expected: "*4\r\n$5\r\nflags\r\n*1\r\n$2\r\non\r\n$9\r\npasswords\r\n*0\r\n",
		},
		{
			name:     "ACL GETUSER non-existent user",
			commands: []string{"ACL", "GETUSER", "nonexistent"},
			initialUser: &types.User{
				Name:      "default",
				Flags:     []string{"on"},
				Passwords: []string{},
			},
			expected: "$-1\r\n",
		},
		{
			name:     "ACL SETUSER default >newpassword (add password, remove nopass)",
			commands: []string{"ACL", "SETUSER", "default", ">newpassword"},
			initialUser: &types.User{
				Name:      "default",
				Flags:     []string{"on", "nopass"}, // Starts with nopass
				Passwords: []string{},
			},
			expected: "+OK\r\n",
			expectedUser: func() *types.User {
				password := "newpassword"
				hash := sha256.Sum256([]byte(password))
				passwordHash := hex.EncodeToString(hash[:])
				return &types.User{
					Name:      "default",
					Flags:     []string{"on"}, // nopass should be removed
					Passwords: []string{passwordHash},
				}
			}(),
		},
		{
			name:     "ACL SETUSER non-existent user",
			commands: []string{"ACL", "SETUSER", "nonexistent", ">password"},
			initialUser: &types.User{
				Name:      "default",
				Flags:     []string{"on"},
				Passwords: []string{},
			},
			expected: "-ERR User not found\r\n",
			expectedUser: &types.User{ // User should not be modified
				Name:      "default",
				Flags:     []string{"on"},
				Passwords: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset defaultUser to its initial state for each test that modifies it
			userCopy := &types.User{
				Name:      tt.initialUser.Name,
				Flags:     append([]string{}, tt.initialUser.Flags...),
				Passwords: append([]string{}, tt.initialUser.Passwords...),
			}

			result := handlers.Acl(tt.commands, userCopy, userMutex)
			if result != tt.expected {
				t.Errorf("For test case %s: Expected result %q, got %q", tt.name, tt.expected, result)
			}

			if tt.expectedUser != nil {
				// Compare the modified user state
				if userCopy.Name != tt.expectedUser.Name {
					t.Errorf("For test case %s: Expected user name %q, got %q", tt.name, tt.expectedUser.Name, userCopy.Name)
				}
				if !compareStringSlices(userCopy.Flags, tt.expectedUser.Flags) {
					t.Errorf("For test case %s: Expected user flags %v, got %v", tt.name, tt.expectedUser.Flags, userCopy.Flags)
				}
				if !compareStringSlices(userCopy.Passwords, tt.expectedUser.Passwords) {
					t.Errorf("For test case %s: Expected user passwords %v, got %v", tt.name, tt.expectedUser.Passwords, userCopy.Passwords)
				}
			}
		})
	}
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

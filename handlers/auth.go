package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"

	"github.com/AYGA2K/dictago/types"
)

func Auth(commands []string, defaultUser types.User) string {
	if len(commands) < 3 {
		return "-ERR wrong number of arguments for 'auth' command\r\n"
	}

	username := strings.ToLower(commands[1])
	password := commands[2]

	if username != "default" {
		return "-WRONGPASS invalid username-password pair or user is disabled.\r\n"
	}

	// Check if user has no passwords (nopass flag is present)
	hasNopass := slices.Contains(defaultUser.Flags, "nopass")

	// If user has nopass flag and no passwords, authentication always succeeds
	if hasNopass && len(defaultUser.Passwords) == 0 {
		return "+OK\r\n"
	}

	// Hash the provided password
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	// Check if the hash matches any stored password
	if slices.Contains(defaultUser.Passwords, passwordHash) {
		return "+OK\r\n"
	}

	return "-WRONGPASS invalid username-password pair or user is disabled.\r\n"
}

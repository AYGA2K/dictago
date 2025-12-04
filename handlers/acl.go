package handlers

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/AYGA2K/dictago/types"
)

func Acl(commands []string, defaultUser *types.User, userMutex *sync.Mutex) string {
	if len(commands) == 0 {
		return ""
	}

	cmd := strings.ToUpper(commands[1])

	switch cmd {
	case "WHOAMI":
		return "$7\r\ndefault\r\n"

	case "GETUSER":
		if len(commands) < 2 {
			return "-ERR wrong number of arguments\r\n"
		}
		user := strings.ToLower(commands[2])
		if user == "default" {
			return formatUserResponse(defaultUser)
		}
		return "$-1\r\n"
	case "SETUSER":
		if len(commands) < 4 {
			return "-ERR wrong number of arguments\r\n"
		}
		user := strings.ToLower(commands[2])
		if user == "default" {
			userMutex.Lock()
			defer userMutex.Unlock()
			return handleSetUser(commands[3:], defaultUser)
		}
		return "-ERR User not found\r\n"

	default:
		return "-ERR unknown subcommand\r\n"
	}
}
func handleSetUser(rules []string, defaultUser *types.User) string {
	for _, rule := range rules {
		if strings.HasPrefix(rule, ">") {
			// Add password
			password := rule[1:]
			hash := sha256.Sum256([]byte(password))
			hashStr := fmt.Sprintf("%x", hash)

			// Store the password hash
			defaultUser.Passwords = append(defaultUser.Passwords, hashStr)

			// Remove nopass flag if present
			newFlags := []string{}
			for _, flag := range defaultUser.Flags {
				if flag != "nopass" {
					newFlags = append(newFlags, flag)
				}
			}
			defaultUser.Flags = newFlags
		}
	}
	return "+OK\r\n"
}

func formatUserResponse(user *types.User) string {
	// Format flags array
	flagsResp := fmt.Sprintf("*%d\r\n", len(user.Flags))
	for _, flag := range user.Flags {
		flagsResp += fmt.Sprintf("$%d\r\n%s\r\n", len(flag), flag)
	}

	// Format passwords array
	passwordsResp := fmt.Sprintf("*%d\r\n", len(user.Passwords))
	for _, pass := range user.Passwords {
		passwordsResp += fmt.Sprintf("$%d\r\n%s\r\n", len(pass), pass)
	}

	return fmt.Sprintf("*4\r\n$5\r\nflags\r\n%s$9\r\npasswords\r\n%s", flagsResp, passwordsResp)
}

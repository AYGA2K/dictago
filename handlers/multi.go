package handlers

import "github.com/AYGA2K/dictago/types"

func Multi(client *types.Client) string {
	client.InMulti = true
	return "+OK\r\n"
}

func Discard(client *types.Client) string {
	if !client.InMulti {
		return "-ERR DISCARD without MULTI\r\n"
	}
	client.InMulti = false
	client.Commands = nil
	return "+OK\r\n"
}

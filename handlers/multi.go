package handlers

import "github.com/AYGA2K/dictago/types"

func Multi(client *types.Client) string {
	client.InMulti = true
	return "+OK\r\n"
}

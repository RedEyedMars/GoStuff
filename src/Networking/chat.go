package Networking

import (
	"strings"
)

func setupChatCommands(registry *ClientRegistry) {

	commands["chat_msg"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		registry.SendMsg(chl, ConstructMessage("chat_msg", msg, chl, user))
	}
	commands["collect_channels"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		var channels []string
		for channel := range c.channels {
			channels = append(channels, channel)
		}
		c.send <- []byte("{channel_names}" + strings.Join(channels, "::"))
	}
}

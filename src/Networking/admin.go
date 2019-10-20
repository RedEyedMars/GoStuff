package Networking

import "Logger"

func setupAdminCommands(registry *ClientRegistry) {

	commands["/mode"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		switch string(msg) {
		case "admin":
			mode = adminPasswordRequired
			c.send <- []byte("{admin_msg}Enter password:")
		default:
			mode = client
		}
		if mode == client {
			Logger.Verbose <- Logger.Msg{"Client.HandleMessages", "Client"}
		} else {
			Logger.Verbose <- Logger.Msg{"Client.HandleMessages", "Ask for Password"}
		}
	}
	commands["/"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		if mode == adminPasswordRequired {
			if string(msg) == adminPassword {
				c.send <- []byte("{admin_msg}Access Granted!")
				mode = admin
			} else {
				c.send <- []byte("{admin_msg}Access Denied!")
				mode = client
			}
		} else {
			if mode == admin {
				HandleAdminCommand(string(msg))
			}
		}
	}
}

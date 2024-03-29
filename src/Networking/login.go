package Networking

import (
	"Events"
	"Logger"
	"crypto/sha256"
	"databasing"
	"fmt"
	"strings"
)

func setupLoginCommands(registry *ClientRegistry) {

	commands["attempt_login"] = func(c *Client, msg []byte, cln []byte, user []byte) {
		hash := sha256.New()
		hash.Write([]byte(adminPassword))
		hash.Write(msg)
		pwdAsString := fmt.Sprintf("%x", hash.Sum(nil)[:])
		if member := <-databasing.RequestMember("ByPwd", pwdAsString); member != nil {
			databasing.AddMemberToMaps(member)
			c.name = member.Name
			Events.FuncEvent("Networking.attempt_login.setupChannel", c.setupChannels)
			c.send <- []byte(fmt.Sprintf("{login_successful;;%s}", member.Name))
			Logger.Verbose <- Logger.Msg{"Login successful"}
		} else {
			c.send <- []byte("{login_failed}Credentials not accepted, either check your password or your username!")
			Logger.Verbose <- Logger.Msg{"Login failed"}
		}
	}
	commands["attempt_signup"] = func(c *Client, msg []byte, cln []byte, user []byte) {
		split := strings.Split(string(msg), ",")
		username, pwd := split[0], split[1]
		if member := <-databasing.RequestMember("ByName", username); member != nil {
			c.send <- []byte("{signup_failed}Username taken!")
		} else {
			Logger.Verbose <- Logger.Msg{"No member found; good!"}
			hash := sha256.New()
			hash.Write([]byte(adminPassword))
			hash.Write([]byte(pwd))
			pwdAsString := fmt.Sprintf("%x", hash.Sum(nil)[:])
			if member := <-databasing.RequestMember("ByPwd", pwdAsString); member == nil {
				member := databasing.NewMemberFull(username)
				Events.GoFuncEvent("client.Signup.AddMember", func() {
					databasing.RequestAction("Members", "Add", member, pwdAsString)
				})
				c.name = member.Name
				<-databasing.RequestAction("Channels", "AddMember", "general", member.Name)
				<-databasing.RequestAction("Channels", "AddMember", "private", member.Name)

				Events.FuncEvent("Networking.attempt_login.setupChannel", c.setupChannels)
				c.send <- []byte(fmt.Sprintf("{signup_successful;;%s}", member.Name))
				Logger.Verbose <- Logger.Msg{"Signup successful"}
			} else {
				c.send <- []byte("{login_failed}Credentials not accepted, try a different password and username!")
			}
		}
	}

	commands["attempt_logout"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		c.send <- []byte("{logout_successful}")
		c.name = "_none_"
		c.channels = make(map[string]*databasing.ClientChannel)
	}
}

func (c *Client) setupChannels() {
	for channel := range databasing.RequestChannelsByName("ByMember", c.name) {
		if channel != nil {
			Logger.Verbose <- Logger.Msg{"Add channel:" + channel.Channel.Name}
			c.channels[channel.Channel.Name] = channel
			channel.Channel.NewClient <- c.send
		}
	}
}

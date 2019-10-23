package Networking

import (
	"Events"
	"databasing"
	"strings"
	"time"
)

func mapKeys(myMap map[string]time.Time) []string {
	keys := make([]string, len(myMap))

	i := 0
	for k := range myMap {
		keys[i] = k
		i++
	}
	return keys
}
func setupChatCommands(registry *ClientRegistry) {

	commands["chat_msg"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		registry.SendMsg(chl, ConstructMessage("chat_msg", msg, chl, user))
		if chl != nil {
			channel_name := string(chl)
			if _, ok := databasing.Channels[channel_name]; ok {

				if user == nil {
					databasing.RequestChatMsgAction("AddMsg0Res", string(msg), string(c.name), channel_name, time.Now().Format("2006-01-02 15:04:05"))
				} else {
					databasing.RequestChatMsgAction("AddMsg0Res", string(msg), string(user), channel_name, time.Now().Format("2006-01-02 15:04:05"))
				}
			}

		}
	}
	commands["collect_channels"] = func(c *Client, msg []byte, chl []byte, user []byte) {
		c.send <- []byte("{channel_names}" + strings.Join(mapKeys(c.channels), "::"))

		for channel, ts := range c.channels {
			Events.GoFuncEvent("networking.chat.setupChatCommands.SendMsgs", func() {

				var newTs *time.Time
				for chatMsg := range databasing.RequestChatMsg("OnChannel", channel, ts) {
					newTs = chatMsg.Timestamp
					c.send <- chatMsg.ToByte()
				}
				if newTs != nil {
					c.channels[channel] = *newTs
				}
				lastTimestamp := <-databasing.RequestTimestamp("Last", channel)
				if lastTimestamp == nil || newTs == nil || newTs.Equal(*lastTimestamp) {
					c.send <- []byte("{channel_up_to_date::" + channel + "}")
				}
			})
		}
	}
}

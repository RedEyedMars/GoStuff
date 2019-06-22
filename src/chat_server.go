package main

import (
	"Events"
	"Networking"
	"common_chat"
)

func Run() {
	defer common_chat.MainEnd()
	var webClient <-chan bool
	Events.FuncEvent("Networking.StartWebClient", func() { webClient = Networking.StartWebClient() })
	<-webClient
}
func main() {
	common_chat.MainStart("main.Run", Run)
}

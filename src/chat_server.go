package main

import (
	"Events"
	"Networking"
	"common_chat"
	"databasing"
	"os"
)

func Run() {
	defer common_chat.MainEnd()
	var webClient <-chan bool
	Events.FuncEvent("Networking.StartWebClient", func() { webClient = Networking.StartWebClient() })
	<-webClient
}
func main() {
	args := os.Args
	if len(args) <= 1 {
		common_chat.MainStart("main.Run", Run)
	} else {
		switch args[1] {
		case "chat_service":
			common_chat.MainStart("main.Run", Run)
		case "setup_database":
			common_chat.MainStart("databasing.Setup", databasing.Start)
		}
	}
}

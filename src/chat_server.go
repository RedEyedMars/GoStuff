package main

import (
	"Events"
	"Networking"
	"common_chat"
	"databasing"
	"math/rand"
	"os"
	"time"
)

func Run(Shutdown chan bool) {
	Events.GoFuncEvent("Networking.StartWebClient", func() {
		Networking.StartWebClient(Shutdown)
	})
}
func main() {

	rand.Seed(time.Now().UTC().UnixNano())
	args := os.Args
	if len(args) <= 1 {
		common_chat.MainStart("main.Run", Run, Networking.HandleAdminCommand, Networking.End)
	} else {
		switch args[1] {
		case "chat_service":
			common_chat.MainStart("main.Run", Run, Networking.HandleAdminCommand, Networking.End)
		case "setup_database":
			common_chat.MainStart("databasing.Setup", databasing.Run, databasing.HandleAdminCommand, databasing.End)
		}
	}
}

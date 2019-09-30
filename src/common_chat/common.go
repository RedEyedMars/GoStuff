package common_chat

import (
	"Config"
	"Events"
	"Logger"
	"bufio"
	"os"
)

func Start() {
	Config.Setup()

	Logger.Start()
}

func Close() {
	Logger.Close()
}

func MainStart(name string, f func(chan bool), adminCommand func(string), end func()) {
	Start()
	Shutdown := make(chan bool, 1)
	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if text[:4] == "exit" {
				Shutdown <- true
				break
			} else {
				adminCommand(text)
			}
		}
	}()
	Events.DoneFuncEvent(name, f, Shutdown)
	<-Shutdown
	end()
	Close()
}

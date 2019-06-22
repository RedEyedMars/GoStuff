package common_chat

import (
	"Config"
	"Events"
	"Logger"
)

func Start() {
	Config.Setup()

	Logger.Start()
}

var done chan bool

func Close() {
	Logger.Close()
}

func MainStart(name string, f func()) {
	done = make(chan bool, 1)
	Start()
	Events.FuncEvent(name, f)
	<-done
	Close()
}
func MainEnd() {
	done <- true
}

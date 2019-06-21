package main

import (
	"Config"
	"Events"
	"Logger"
	"Networking"
)

func Start() {
	Config.Setup()

	Logger.Start()
}

var done chan bool

func Run() {
	defer func() { done <- true }()
	var webClient <-chan bool
	Events.FuncEvent("Networking.StartWebClient", func() { webClient = Networking.StartWebClient() })
	<-webClient
}

func Close() {
	Logger.Close()
}
func main() {
	done = make(chan bool, 1)
	Start()
	Events.FuncEvent("main.Run", Run)
	<-done
	Close()
}

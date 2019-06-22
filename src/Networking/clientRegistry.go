package Networking

import (
	"Events"
	"Logger"
)

type ClientRegistry struct {
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
	ipNames    map[string]string
	broadcast  chan []byte
}

func newRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (h *ClientRegistry) run() {
	Logger.Verbose <- Logger.Msg{"ClientRegistry.Run", "Begin"}
	defer func() { Logger.Verbose <- Logger.Msg{"ClientRegistry.Run", "Finish"} }()
	for {
		select {
		case client := <-h.register:
			Events.FuncEvent("ClientRegistry.Register", func() {
				h.clients[client] = true
				if name, present := h.ipNames[client.ip]; present {
					client.name = name
				}
			})
		case client := <-h.unregister:
			Events.FuncEvent("ClientRegistry.Unregister", func() {
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			})
		case message := <-h.broadcast:
			Events.FuncEvent("ClientRegistry.Broadcast", func() {
				for client := range h.clients {
					select {
					case client.send <- SanatizeMessage(message):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			})

		}
	}
}

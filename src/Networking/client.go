package Networking

import (
	"Events"
	"Logger"
	"bytes"
	"crypto/sha256"
	"databasing"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	ip string

	port int

	name string

	// Buffered channel of outbound messages.
	send chan []byte
	// Registered clients.

	// Inbound messages from the clients.
	handle chan []byte
}

func newClient(conn *websocket.Conn) *Client {
	ip, port := GetIPFromAddress(conn.RemoteAddr().String())
	return &Client{
		conn:   conn,
		ip:     ip,
		port:   port,
		name:   "_none_",
		send:   make(chan []byte, 256),
		handle: make(chan []byte)}
}

// readMessages pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readMessages(registry *ClientRegistry) {
	defer func() {
		registry.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "Unexpected"}
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.handle <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(registry *ClientRegistry, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Warning <- Logger.Msg{err.Error(), "Networking.serveWs"}
		return
	}
	client := newClient(conn)
	registry.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	Events.GoFuncEvent("Networking.WriteMessages", client.writeMessages)
	Events.GoFuncEvent("Networking.ReadMessages", func() { client.readMessages(registry) })
	Events.GoFuncEvent("Networking.HandleMesssages", func() { client.handleMessages(registry) })
}

const (
	client                int = 0
	adminPasswordRequired     = 1
	admin                     = 2

	adminPassword = "SuperSecurePassword"
)

var mode = client
var adminCommands map[string]Events.Event
var adminArgs []string

func (c *Client) handleMessages(registry *ClientRegistry) {
	for {
		select {
		case message := <-c.handle:
			Logger.VeryVerbose <- Logger.Msg{string(message), "Receive"}

			switch command, msg := DifferentiateMessage(message); command {
			case "chat_msg":
				registry.broadcast <- message
			case "new_connection":
				//registry.broadcast <- []byte("{new_connection}" + c.conn.LocalAddr().String() + "::" + c.conn.RemoteAddr().String())
				//registry.broadcast <- message
			case "attempt_login":
				Events.GoFuncEvent("client.AttemptLogin", func() {
					hash := sha256.New()
					hash.Write([]byte(adminPassword))
					hash.Write(msg)
					if member := <-databasing.RequestMember("ByPwd", string(hash.Sum(nil))); member != nil {
						c.name = member.Name
						c.send <- []byte("{login_successful}" + member.Name)
					} else {
						c.send <- []byte("{login_failed}Credentials not accepted, either check your password or your username!")
					}
				})
			case "attempt_signup":
				Events.GoFuncEvent("client.AttemptSignup", func() {
					split := strings.Split(string(msg), ",")
					username, pwd := split[0], split[1]
					if member := <-databasing.RequestMember("ByName", username); member != nil {
						c.send <- []byte("{signup_failed}Username taken!")
					} else {
						hash := sha256.New()
						hash.Write([]byte(adminPassword))
						hash.Write([]byte(pwd))
						pwdAsString := string(hash.Sum(nil))
						if member := <-databasing.RequestMember("ByPwd", pwdAsString); member == nil {
							member := databasing.NewMemberFull(username)
							Events.GoFuncEvent("client.Signup.AddMember", func() {
								databasing.AddMemberToMaps(member)
								databasing.RequestMemberAction("Add", member, pwdAsString)
							})
							c.name = member.Name
							c.send <- []byte("{signup_successful}" + member.Name)
						} else {
							c.send <- []byte("{login_failed}Credentials not accepted, try a different password and username!")
						}
					}
				})
			case "/mode":
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

			case "/pwd":
				if mode == adminPasswordRequired && string(msg) == adminPassword {
					c.send <- []byte("{admin_msg}Access Granted!")
					mode = admin
				} else {
					c.send <- []byte("{admin_msg}Access Denied!")
					mode = client
				}
			case "/":
				if mode == admin {
					HandleAdminCommand(string(msg))
				}
			}
			//default:
			//	Logger.VeryVerbose <- Logger.Msg{"HandleMessages.Unregister"}
			//	registry.unregister <- c
		}
	}
}

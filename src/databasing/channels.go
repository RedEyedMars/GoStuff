package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"time"
)

type Channel struct {
	Name         string
	Members      map[string]*Member
	OrderID      int
	Send         chan []byte
	NewClient    chan chan []byte
	RemoveClient chan chan []byte
}
type DBChannelResponse struct {
	chl       chan *Channel
	assembler func(*sql.Rows) *Channel
}

func (mr *DBChannelResponse) send(result *sql.Rows) {
	mr.chl <- mr.assembler(result)
}
func (mr *DBChannelResponse) close() {
	close(mr.chl)
}

type ClientChannel struct {
	Channel   *Channel
	LastKnown time.Time
}
type DBClientChannelResponse struct {
	chl       chan *ClientChannel
	assembler func(*sql.Rows) *ClientChannel
}

func (mr *DBClientChannelResponse) send(result *sql.Rows) {
	mr.chl <- mr.assembler(result)
}
func (mr *DBClientChannelResponse) close() {
	close(mr.chl)
}

func NewChannel(name string, id int) *Channel {
	channel := &Channel{
		Name:         name,
		Members:      make(map[string]*Member),
		OrderID:      id,
		Send:         make(chan []byte),
		NewClient:    make(chan chan []byte),
		RemoveClient: make(chan chan []byte),
	}
	channel.HookUp()
	channel.AddChannelToMaps()
	return channel
}
func (channel *Channel) HookUp() {
	channel.Send = make(chan []byte, 16)
	clients := make(map[chan []byte]bool)
	Events.GoFuncEvent("databasing.Channel.HookUp."+channel.Name, func() {
		for {
			select {
			case client := <-channel.NewClient:
				clients[client] = true
			case client := <-channel.RemoveClient:
				delete(clients, client)
			case msg := <-channel.Send:
				for client := range clients {
					client <- msg
				}
			}
		}
	})
}

func LoadAllChannels() {
	request := RequestChannel("AllNames")
	channel := <-request
	for channel != nil {
		channel = <-request
	}
}

func (channel *Channel) AddChannelToMaps() {
	Logger.VeryVerbose <- Logger.Msg{"Add Channel:" + channel.Name}
	Channels[channel.Name] = channel
}

func SetupChannels(db *sql.DB) {
	defineQuery(db, "Channels_All", `SELECT channel_name,member_name,id FROM channels_user_info ;`)
	defineQuery(db, "Channels_AllNames", `SELECT channel_name,id FROM channels_user_info ;`)

	defineQuery(db, "Channels_ByMember", `SELECT channel_name,last_known FROM channels_user_info WHERE member_name=? ;`)
	defineQuery(db, "Channels_Channels", `SELECT channel_name FROM channels_user_info;`)

	defineQuery(db, "Channels_AddMember", `INSERT INTO channels_user_info VALUES (?,?,?,NULL);`)
}
func RequestChannel(name string, args ...interface{}) <-chan *Channel {
	response := make(chan *Channel, 1)
	queries <- &DBQueryResponse{
		query: "Channels_" + name,
		args:  args,
		sender: &DBChannelResponse{
			chl:       make(chan *Channel, 1),
			assembler: parseChannel,
		},
	}
	return response
}
func RequestChannelsByName(name string, args ...interface{}) <-chan *ClientChannel {
	response := make(chan *ClientChannel, 1)
	queries <- &DBQueryResponse{
		query: "Channels_" + name,
		args:  args,
		sender: &DBClientChannelResponse{
			chl:       make(chan *ClientChannel, 1),
			assembler: parseClientChannel,
		},
	}
	return response
}

func parseChannel(rows *sql.Rows) *Channel {
	var (
		name string
		id   int
	)
	if err := rows.Scan(&name, &id); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
	}
	return NewChannel(name, id)
}
func parseClientChannel(rows *sql.Rows) *ClientChannel {
	var (
		name       string
		last_known time.Time
	)

	if err := rows.Scan(&name, &last_known); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
	}
	return &ClientChannel{Channels[name], last_known}
}

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

type ClientChannel struct {
	Channel   *Channel
	LastKnown time.Time
}

type DBChannelResponse struct {
	Query    func() (*sql.Rows, error)
	Channels chan *Channel
}
type DBClientChannelResponse struct {
	Query    func() (*sql.Rows, error)
	Channels chan *ClientChannel
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
func NewChannelResponse(name string, arg ...interface{}) *DBChannelResponse {
	return NewChannelResponseArr(name, arg)
}
func NewChannelResponseArr(name string, args []interface{}) *DBChannelResponse {
	return &DBChannelResponse{
		Query:    func() (*sql.Rows, error) { return dbQueries["Channels_"+name].Query(args...) },
		Channels: make(chan *Channel, 1)}
}
func NewClientChannelResponseArr(name string, args []interface{}) *DBClientChannelResponse {
	return &DBClientChannelResponse{
		Query:    func() (*sql.Rows, error) { return dbQueries["Channels_"+name].Query(args...) },
		Channels: make(chan *ClientChannel, 1)}
}
func NewChannelActionArr(name string, args []interface{}) *DBActionResponse {
	return &DBActionResponse{
		Exec:       func() (sql.Result, error) { return dbQueries["Channels_"+name].Exec(args...) },
		Successful: make(chan bool, 1)}
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
	request := NewChannelResponseArr(name, args)
	ChannelRequests <- request
	return request.Channels
}
func RequestChannelsByName(name string, args ...interface{}) <-chan *ClientChannel {
	request := NewClientChannelResponseArr(name, args)
	ChannelNamesRequests <- request
	return request.Channels
}
func RequestChannelAction(name string, args ...interface{}) <-chan bool {
	var request = NewChannelActionArr(name, args)
	ActionRequests <- request
	return request.Successful
}
func (chs *DBChannelResponse) ParseNew(rows *sql.Rows) {
	for rows.Next() {
		var (
			name string
			id   int
		)
		if err := rows.Scan(&name, &id); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
		}
		channel := NewChannel(name, id)
		chs.Channels <- channel
	}
	close(chs.Channels)
}
func (chs *DBClientChannelResponse) ParseNames(rows *sql.Rows) {
	for rows.Next() {
		var (
			name       string
			last_known time.Time
		)

		if err := rows.Scan(&name, &last_known); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
		}

		chs.Channels <- &ClientChannel{Channels[name], last_known}
	}
	close(chs.Channels)
}

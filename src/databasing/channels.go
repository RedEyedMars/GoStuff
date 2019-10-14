package databasing

import (
	"Events"
	"Logger"
	"database/sql"
)

type Channel struct {
	Name      string
	Members   map[string]*Member
	OrderID   int
	Send      chan []byte
	NewClient chan chan []byte
}

type DBChannelResponse struct {
	Query    func() (*sql.Rows, error)
	Channels chan *Channel
}

func NewChannel(name string, id int) *Channel {
	channel := &Channel{
		Name:      name,
		Members:   make(map[string]*Member),
		OrderID:   id,
		Send:      make(chan []byte),
		NewClient: make(chan chan []byte),
	}
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
			case msg := <-channel.Send:
				for client := range clients {
					select {
					case client <- msg:
					default:
						delete(clients, client)
						close(client)
					}
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
func NewChannelActionArr(name string, args []interface{}) *DBActionResponse {
	return &DBActionResponse{
		Exec:       func() (sql.Result, error) { return dbQueries["Channels_"+name].Exec(args...) },
		Successful: make(chan bool, 1)}
}

func LoadAllChannels() {
	for channel := range RequestChannel("AllNames") {
		channel.HookUp()
		channel.AddChannelToMaps()
	}
}

func (channel *Channel) AddChannelToMaps() {
	Logger.VeryVerbose <- Logger.Msg{"Add Channel:" + channel.Name}
	Channels[channel.Name] = channel
}

func SetupChannels(db *sql.DB) {
	defineQuery(db, "Channels_All", `SELECT channel_name,member_name,id FROM channels_names ;`)
	defineQuery(db, "Channels_AllNames", `SELECT channel_name,id FROM channels_names ;`)

	defineQuery(db, "Channels_ByMember", `SELECT channel_name FROM channels_names WHERE member_name=? ;`)
	defineQuery(db, "Channels_Channels", `SELECT channel_name FROM channels_names;`)

	defineQuery(db, "Channels_AddMember", `INSERT INTO channels_names VALUES (?,?,NULL);`)
}
func RequestChannel(name string, args ...interface{}) <-chan *Channel {
	request := NewChannelResponseArr(name, args)
	ChannelRequests <- request
	return request.Channels
}
func RequestChannelsByName(name string, args ...interface{}) <-chan *Channel {
	request := NewChannelResponseArr(name, args)
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
func (chs *DBChannelResponse) ParseNames(rows *sql.Rows) {
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
		}

		chs.Channels <- Channels[name]
	}
	close(chs.Channels)
}

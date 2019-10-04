package databasing

import (
	"Logger"
	"database/sql"
)

type Channel struct {
	Name    string
	Members map[string]*Member
	OrderID int
}

type DBChannelResponse struct {
	Query    func() (*sql.Rows, error)
	Channels chan *Channel
}

func NewChannelResponse(name string, arg ...string) *DBChannelResponse {
	return NewChannelResponseArr(name, arg)
}
func NewChannelResponseArr(name string, args []string) *DBChannelResponse {
	return &DBChannelResponse{
		Query:    func() (*sql.Rows, error) { return dbQueries["Channels_"+name].Query(args) },
		Channels: make(chan *Channel, 1)}
}

func SetupChannels(db *sql.DB) {
	defineQuery(db, "Channels_All", `SELECT channel_name,member_name,id FROM channels_names ;`)
	defineQuery(db, "Channels_AllNames", `SELECT channel_name,id FROM channels_names ;`)

	defineQuery(db, "Channels_Members", `SELECT member_name FROM channels_names WHERE chanel_name=? ;`)
	defineQuery(db, "Members_Channels", `SELECT channel_name FROM channels_names WHERE member_name=? ;`)
	defineQuery(db, "Channels_Channels", `SELECT channel_name FROM channels_names;`)
}
func RequestChannel(name string, args ...string) <-chan *Channel {
	request := NewChannelResponseArr(name, args)
	ChannelRequests <- request
	return request.Channels
}
func RequestChannelsByName(name string, args ...string) <-chan *Channel {
	request := NewChannelResponseArr(name, args)
	ChannelNamesRequests <- request
	return request.Channels
}
func (chs *DBChannelResponse) ParseNew(rows *sql.Rows) {
	var channel *Channel
	channel = nil
	for rows.Next() {
		var (
			name string
			id   int
		)
		if err := rows.Scan(&name, &id); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
		}
		chs.Channels <- channel
	}
	close(chs.Channels)
}
func (chs *DBChannelResponse) ParseNames(rows *sql.Rows) {
	for rows.Next() {
		var (
			name string
			id   int
		)
		if err := rows.Scan(&name, &id); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.channels.Parse"}
		}

		chs.Channels <- Channels[name]
	}
	close(chs.Channels)
}

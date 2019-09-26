package databasing

import (
	"Logger"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Channel struct {
	Name    string
	Members map[string]*Member
	OrderID int
}

type DBChannelResponse struct {
	Query    func() string
	Channels chan *Channel
}

func NewChannelResponse(name string, arg ...string) *DBChannelResponse {
	return NewChannelResponseArr(name, arg)
}
func NewChannelResponseArr(name string, arg []string) *DBChannelResponse {
	name = "Channels_" + name
	switch dbQueryArgumentLength[name] {
	case 0:
		return &DBChannelResponse{
			Query:    func() string { return dbQueries[name] },
			Channels: make(chan *Channel, 1)}
	case 1:
		return &DBChannelResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0]) },
			Channels: make(chan *Channel, 1)}
	case 2:
		return &DBChannelResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1]) },
			Channels: make(chan *Channel, 1)}
	case 3:
		return &DBChannelResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1], arg[2]) },
			Channels: make(chan *Channel, 1)}
	default:
		Logger.Error <- Logger.ErrMsg{Err: errors.New(name + "has to many arguments:" + strconv.Itoa(dbQueryArgumentLength[name])), Status: "databasing.channels.NewMemberResponse"}
		return nil
	}
}

func SetupChannels() {
	defineQuery("Channels_All", `SELECT channel_name,member_name,id FROM channels_names ;`, 0)
	defineQuery("Channels_AllNames", `SELECT channel_name,id FROM channels_names ;`, 0)

	defineQuery("Channels_Members", `SELECT member_name FROM channels_names WHERE chanel_name='%s' ;`, 1)
	defineQuery("Members_Channels", `SELECT channel_name FROM channels_names WHERE member_name='%s' ;`, 1)
	defineQuery("Channels_Channels", `SELECT channel_name FROM channels_names ;`, 1)
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

package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type ChatMsg struct {
	Text      string
	Channel   string
	Sender    string
	Resources string
	Timestamp *time.Time
	OrderID   int
}

func (cm *ChatMsg) LoadResources() {
	for i := 3; i < len(cm.Resources); i += 3 {
		if _, present := LoadedResources[cm.Resources[i-3:i]]; !present {

			Events.GoFuncEvent("databasing.chatmsgs.LoadResource", func() {
				LoadedResources[cm.Resources[i-3:i]] = <-RequestResource("ByAbv", cm.Resources[i-3:i])
			})
		}
	}
}

type DBChatMsgResponse struct {
	Query    func() string
	ChatMsgs chan ChatMsg
}

func NewChatMsgResponse(name string, arg ...string) *DBChatMsgResponse {
	return NewChatMsgResponseArr(name, arg)
}
func NewChatMsgResponseArr(name string, arg []string) *DBChatMsgResponse {
	name = "ChatMsg_" + name
	switch dbQueryArgumentLength[name] {
	case 0:
		return &DBChatMsgResponse{
			Query:    func() string { return dbQueries[name] },
			ChatMsgs: make(chan ChatMsg, 1)}
	case 1:
		return &DBChatMsgResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0]) },
			ChatMsgs: make(chan ChatMsg, 1)}
	case 2:
		return &DBChatMsgResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1]) },
			ChatMsgs: make(chan ChatMsg, 1)}
	case 3:
		return &DBChatMsgResponse{
			Query:    func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1], arg[2]) },
			ChatMsgs: make(chan ChatMsg, 1)}
	default:
		Logger.Error <- Logger.ErrMsg{Err: errors.New(name + "has to many arguments:" + strconv.Itoa(dbQueryArgumentLength[name])), Status: "databasing.ChatMsg.NewChatMsgResponse"}
		return nil
	}
}

func SetupChatMsgs() {
	defineQuery("ChatMsg_RecentOnChannel", `SELECT msg,sender,channel,time_sent,number_of_resources,resources,id FROM made_orders WHERE channel = '%s' AND timestamp >= NOW() - INTERVAL 24 HOUR  LIMIT 16 ;`, 1)
	defineQuery("ChatMsg_ByIdOnChannel", `SELECT msg,sender,channel,time_sent,number_of_resources,resources,id FROM made_orders WHERE id < %s AND channel = '%s'  LIMIT 16 ;`, 2)
}
func RequestChatMsg(name string, args ...string) <-chan ChatMsg {
	request := NewChatMsgResponseArr(name, args)
	ChatMsgRequests <- request
	return request.ChatMsgs
}
func (cm *DBChatMsgResponse) Parse(rows *sql.Rows) {
	for rows.Next() {
		var (
			text              string
			channel           string
			sender            string
			numberOfResources int
			resources         string
			timestamp         *time.Time
			orderId           int
		)
		if err := rows.Scan(&text, &channel, &sender, &numberOfResources, &resources, &timestamp, &orderId); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.chatmsgs.Parse"}
		}

		cm.ChatMsgs <- ChatMsg{text, channel, sender, resources, timestamp, orderId}
	}
	close(cm.ChatMsgs)
}

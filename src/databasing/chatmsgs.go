package databasing

import (
	"Events"
	"Logger"
	"database/sql"
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
	Query    func() (*sql.Rows, error)
	ChatMsgs chan ChatMsg
}

func NewChatMsgResponse(name string, arg ...string) *DBChatMsgResponse {
	return NewChatMsgResponseArr(name, arg)
}
func NewChatMsgResponseArr(name string, args []string) *DBChatMsgResponse {
	return &DBChatMsgResponse{
		Query:    func() (*sql.Rows, error) { return dbQueries["ChatMsg_"+name].Query(args) },
		ChatMsgs: make(chan ChatMsg, 1)}
}

func SetupChatMsgs(db *sql.DB) {
	defineQuery(db, "ChatMsg_RecentOnChannel", `SELECT msg,sender,channel,time_sent,number_of_resources,resources,id FROM made_orders WHERE channel = ? AND timestamp >= NOW() - INTERVAL 24 HOUR  LIMIT 16 ;`)
	defineQuery(db, "ChatMsg_ByIdOnChannel", `SELECT msg,sender,channel,time_sent,number_of_resources,resources,id FROM made_orders WHERE id < ? AND channel = ?  LIMIT 16 ;`)
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

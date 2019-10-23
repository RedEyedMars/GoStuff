package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"fmt"
	"time"
)

/**
+------------------------------+------------------+------+-----+---------+----------------+
| Field                        | Type             | Null | Key | Default | Extra          |
+------------------------------+------------------+------+-----+---------+----------------+
| msg                          | text             | NO   | MUL | NULL    |                |
| sender                       | varchar(255)     | NO   |     | NULL    |                |
| channel                      | varchar(255)     | YES  |     | NULL    |                |
| time_sent                    | timestamp        | NO   |     | NULL    |                |
| number_of_required_resources | tinyint(4)       | YES  |     | NULL    |                |
| resources                    | varchar(255)     | YES  |     | NULL    |                |
| id                           | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
+------------------------------+------------------+------+-----+---------+----------------+
**/
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
type DBTimestampResponse struct {
	Query      func() (*sql.Rows, error)
	Timestamps chan *time.Time
}

func NewChatMsgResponse(name string, arg ...interface{}) *DBChatMsgResponse {
	return NewChatMsgResponseArr(name, arg)
}
func NewChatMsgResponseArr(name string, args []interface{}) *DBChatMsgResponse {
	return &DBChatMsgResponse{
		Query:    func() (*sql.Rows, error) { return dbQueries["ChatMsg_"+name].Query(args...) },
		ChatMsgs: make(chan ChatMsg, 1)}
}
func NewTimestampResponseArr(name string, args []interface{}) *DBTimestampResponse {
	return &DBTimestampResponse{
		Query:      func() (*sql.Rows, error) { return dbQueries["Timestamp_"+name].Query(args...) },
		Timestamps: make(chan *time.Time, 1)}
}
func NewChatMsgActionArr(name string, args []interface{}) *DBActionResponse {
	return &DBActionResponse{
		Exec:       func() (sql.Result, error) { return dbQueries["ChatMsg_"+name].Exec(args...) },
		Successful: make(chan bool, 1)}
}

func SetupChatMsgs(db *sql.DB) {
	defineQuery(db, "ChatMsg_RecentOnChannel", `SELECT msg,sender,channel,time_sent,number_of_required_resources,resources,id FROM messages WHERE channel = ? AND time_sent >= NOW() - INTERVAL 24 HOUR  LIMIT 16 ;`)
	defineQuery(db, "ChatMsg_ByIdOnChannel", `SELECT msg,sender,channel,time_sent,number_of_required_resources,resources,id FROM messages WHERE id > ? AND channel = ?  LIMIT 16 ;`)

	defineQuery(db, "ChatMsg_OnChannel", `SELECT msg,sender,channel,time_sent,number_of_required_resources,resources,id FROM messages WHERE channel = ? AND time_sent >= ?  LIMIT 16 ;`)
	defineQuery(db, "Timestamp_Last", `SELECT MAX(id),time_sent FROM messages LIMIT 1`)

	defineQuery(db, "ChatMsg_AddMsg0Res", `INSERT INTO messages VALUES(?, ?, ?, ?, 0, NULL, NULL) ;`)

}
func RequestChatMsg(name string, args ...interface{}) <-chan ChatMsg {
	request := NewChatMsgResponseArr(name, args)
	ChatMsgRequests <- request
	return request.ChatMsgs
}
func RequestTimestamp(name string, args ...interface{}) <-chan *time.Time {
	request := NewTimestampResponseArr(name, args)
	TimestampRequests <- request
	return request.Timestamps
}
func RequestChatMsgAction(name string, args ...interface{}) <-chan bool {
	request := NewChatMsgActionArr(name, args)
	ActionRequests <- request
	return request.Successful
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
			orderID           int
		)
		if err := rows.Scan(&text, &channel, &sender, &numberOfResources, &resources, &timestamp, &orderID); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.chatmsgs.Parse"}
		}

		cm.ChatMsgs <- ChatMsg{text, channel, sender, resources, timestamp, orderID}
	}
	close(cm.ChatMsgs)
}
func (cm *DBTimestampResponse) Parse(rows *sql.Rows) {
	for rows.Next() {
		var (
			timestamp *time.Time
			orderID   int
		)
		if err := rows.Scan(&orderID, &timestamp); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.chatmsgs.ts.Parse"}
		}

		cm.Timestamps <- timestamp
	}
	close(cm.Timestamps)
}
func (cm *ChatMsg) ToByte() []byte {
	return []byte(fmt.Sprintf("{chat_msg::%s;;%s}%s", cm.Channel, cm.Sender, cm.Text))
}

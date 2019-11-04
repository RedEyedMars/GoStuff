package databasing

import (
	"Events"
	"Logger"
	"database/sql"
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
type DBChatMsgResponse struct {
	chl       chan *ChatMsg
	assembler func(*sql.Rows) *ChatMsg
}

func (mr *DBChatMsgResponse) send(result *sql.Rows) {
	mr.chl <- mr.assembler(result)
}
func (mr *DBChatMsgResponse) close() {
	close(mr.chl)
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

func SetupChatMsgs(db *sql.DB) {
	defineQuery(db, "ChatMsg_RecentOnChannel", `SELECT msg,sender,channel,time_sent,number_of_required_resources,resources,id FROM messages WHERE channel = ? AND time_sent >= NOW() - INTERVAL 24 HOUR  LIMIT 16 ;`)
	defineQuery(db, "ChatMsg_ByIdOnChannel", `SELECT msg,sender,channel,time_sent,number_of_required_resources,resources,id FROM messages WHERE id < ? AND channel = ?  LIMIT 16 ;`)
}
func RequestChatMsg(name string, args ...interface{}) <-chan *ChatMsg {
	response := make(chan *ChatMsg, 1)
	queries <- &DBQueryResponse{
		query: "ChatMsg_" + name,
		args:  args,
		sender: &DBChatMsgResponse{
			chl:       response,
			assembler: parseChatMsg,
		},
	}
	return response
}
func parseChatMsg(rows *sql.Rows) *ChatMsg {
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

	return &ChatMsg{text, channel, sender, resources, timestamp, orderId}
}

package databasing

import (
	"Events"
	"Logger"
	"common_chat"
	"database/sql"
	"fmt"
	"log"
)

var dbQueries map[string]string
var dbQueryArgumentLength map[string]int
var ResourceRequests chan *DBResourceResponse

func Setup() {
	dbQueries = make(map[string]string)
	dbQueryArgumentLength = make(map[string]int)
	ResourceRequests = make(chan *DBResourceResponse)
	Events.FuncEvent("databasing.SetupResources", SetupResources())
}

func Start() {
	defer common_chat.MainEnd()
	Logger.Verbose <- Logger.Msg{"Setting up database..."}
	Events.FuncEvent("databasing.StartDatabase", func() {
		dbName := "chat_msg"
		dbEndpoint := "chat-service.c84g8cm4el5a.us-west-2.rds.amazonaws.com"

		dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true", "chat_root", dbPassword, dbEndpoint, dbName)

		// Use db to perform SQL operations on database
		if db, err := sql.Open("mysql", dnsStr); err != nil {
			log.Fatal(err)
		} else {
			defer db.Close()
			if err = db.Ping(); err != nil {
				log.Fatal(err)
			}
			Events.GoFuncEvent("databasing.StartMessageListening", func() { StartMessageListening(db) })
			Logger.Verbose <- Logger.Msg{"Closing database..."}
		}
	})
}
func StartMessageListening(db *sql.DB) {
	for {
		select {
		case request := <-ResourceRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{err, "StartMessageListening.ResourceRequest.Query"}
			} else {
				for rows.Next() {
					var (
						name   string
						source string
						abv    string
					)
					if err := rows.Scan(&name, &source, &abv); err != nil {
						Logger.Error <- Logger.ErrMsg{err, "StartMessageListening.ResourceRequest.Scan"}
					}
					request.Resources <- Resource{name, source, abv}
				}
			}
		default:
			return
		}
	}
}

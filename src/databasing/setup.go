package databasing

import (
	"Events"
	"Logger"
	"common_chat"
	"database/sql"
	"fmt"
	"log"
)

func Setup() {
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
			Logger.Verbose <- Logger.Msg{"Closing database..."}
		}
	})
}

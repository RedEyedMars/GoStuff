package databasing

import (
	"Events"
	"Logger"
	"common_chat"
	"database/sql"
	"log"
)

func Setup() {
	defer common_chat.MainEnd()
	Logger.Verbose <- Logger.Msg{"Setting up database..."}
	Events.FuncEvent("databasing.StartDatabase", func() {
		if db, err := sql.Open("mysql", "user:password@tcp(45.72.131.186:8081)/chat_service"); err != nil {
			log.Fatal(err)
		} else {
			defer db.Close()
			if err = db.Ping(); err != nil {
				log.Fatal(err)
			}
		}
	})
}

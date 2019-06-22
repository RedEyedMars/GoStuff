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
		dbName := "chat-service"
		dbEndpoint := "45.72.131.186:3306"
		//awsRegion := "us-west-2c"
		dbUser := "chat_root"
		awsCreds := "ZXasqw12"
		//authToken, err := rdsutils.BuildAuthToken(dbEndpoint, awsRegion, dbUser, awsCreds)

		// Create the MySQL DNS string for the DB connection
		// user:password@protocol(endpoint)/dbname?<params>
		dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true",
			dbUser, awsCreds, dbEndpoint, dbName,
		)

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

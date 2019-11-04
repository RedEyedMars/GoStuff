package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

/**
+---------------------+
| Tables_in_chat_msg  |
+---------------------+
| channels_names      |
| client_names        |
| messages            |
| resource_allocation |
| resources           |
+---------------------+
**/

var dbQueries map[string]*sql.Stmt

var queries chan dbQuery
var actions chan dbQuery

var LoadedResources map[string]*Resource
var Channels map[string]*Channel

var reSanatizeDatabase *regexp.Regexp
var reIsName *regexp.Regexp

var adminCommands map[string]Events.Event
var adminArgs []string

type sendable interface{}
type dbQuery interface {
	execute()
}
type dbSender interface {
	send(*sql.Rows)
	close()
}
type DBActionResponse struct {
	exec string
	args []interface{}
	chl  chan bool
}
type DBQueryResponse struct {
	query  string
	args   []interface{}
	sender dbSender
}

func (r *DBActionResponse) execute() {
	if result, err := dbQueries[r.exec].Exec(); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.query.action" + r.exec}
	} else {
		if _, err := result.RowsAffected(); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.query.action" + r.exec}
		} else {
			Events.GoFuncEvent("databasing.query.action"+r.exec, func() {
				r.chl <- true
				close(r.chl)
			})
		}

	}
}
func (r *DBQueryResponse) execute() {
	if rows, err := dbQueries[r.query].Query(); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.query.request" + r.query}
	} else {
		Events.GoFuncEvent("databasing.query.request"+r.query, func() {
			for rows.Next() {
				r.sender.send(rows)
			}
			r.sender.close()
		})
	}
}
func RequestAction(mode string, name string, args ...interface{}) <-chan bool {
	response := make(chan bool, 1)
	actions <- &DBActionResponse{
		exec: mode + "_" + name,
		args: args,
		chl:  response,
	}
	return response
}

func makeAdminFunc(argCount uint16, f func(...string)) func() {
	switch argCount {
	case 0:
		return func() { f() }
	case 1:
		return func() {
			if adminArgs != nil && len(adminArgs) > 0 {
				f(adminArgs[0][:len(adminArgs[0])-1])
			}
		}
	case 2:
		return func() {
			if adminArgs != nil && len(adminArgs) > 1 {
				f(adminArgs[0], adminArgs[1][:len(adminArgs[1])-1])
			}
		}
	}
	return func() {}
}
func SetupAdminCommands() {
	if adminCommands == nil {
		adminCommands = make(map[string]Events.Event)
		//adminCommands["exit"] = &Events.Function{Name: "Admin!Exit", Function: func() { Shutdown <- true }}
		adminCommands["addMember"] = &Events.Function{Name: "Admin!AddMember_Full", Function: makeAdminFunc(2,
			func(args ...string) { RequestAction("Members", "Add", NewMemberFull(args[0]).Name, args[1]) })}
		adminCommands["removeMember"] = &Events.Function{Name: "Admin!RemoveMember", Function: makeAdminFunc(1,
			func(args ...string) { RequestAction("Members", "Remove", Members[args[0]].Name) })}
	}
}
func HandleAdminCommand(msg string) bool {
	splice := strings.Split(msg, " ")
	if len(splice) == 1 {
		if command := adminCommands[msg]; command == nil {
			return false
		} else {
			Events.HandleEvent(command)
			return true
		}
	} else {
		if command := adminCommands[splice[0]]; command == nil {
			return false
		} else {
			adminArgs = splice[1:]
			Events.HandleEvent(command)
			return true
		}
	}
}
func Setup() {
	dbQueries = make(map[string]*sql.Stmt)

	SetupAdminCommands()

	queries = make(chan dbQuery, 16)
	actions = make(chan dbQuery, 16)

	LoadedResources = make(map[string]*Resource)
	Channels = make(map[string]*Channel)
	Members = make(map[string]*Member)

	reSanatizeDatabase = regexp.MustCompile(`(\n, \r, \, ', ")`)
	reIsName = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]*`)
}
func defineQuery(db *sql.DB, name string, query string) {
	if stmt, err := db.Prepare(query); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.defineQuery: Failed to define:" + name}
	} else {
		dbQueries[name] = stmt
	}
}
func Run(Shutdown chan bool) {
	Logger.Verbose <- Logger.Msg{"Setting up database..."}
	Events.GoFuncEvent("databasing.Run", func() {
		Events.FuncEvent("databasing.Setup", Setup)
		Events.FuncEvent("databasing.StartDatabase", func() { StartDatabase(Shutdown) })
	})
}

func StartDatabase(Shutdown chan bool) {
	dbUser := "chat_root"
	dbName := "chat_msg"
	dbEndpoint := "chat-service.c84g8cm4el5a.us-west-2.rds.amazonaws.com"

	// Create the MySQL DNS string for the DB connection
	// user:password@protocol(endpoint)/dbname?<params>

	dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPassword, dbEndpoint, dbName)

	// Use db to perform SQL operations on database
	if db, err := sql.Open("mysql", dnsStr); err != nil {
		log.Fatal(err)
	} else {
		if err = db.Ping(); err != nil {
			log.Fatal(err)
		}

		onClose = func() {
			Logger.Verbose <- Logger.Msg{"Closing database..."}
			db.Close()
		}

		Events.FuncEvent("databasing.SetupResources", func() { SetupResources(db) })
		Events.FuncEvent("databasing.SetupChatMsgs", func() { SetupChatMsgs(db) })
		Events.FuncEvent("databasing.SetupMembers", func() { SetupMembers(db) })
		Events.FuncEvent("databasing.SetupChannels", func() { SetupChannels(db) })

		Events.GoFuncEvent("databasing.LoadAllMembers", func() { LoadAllMembers() })
		Events.GoFuncEvent("databasing.LoadAllChannels", func() { LoadAllChannels() })

		Events.FuncEvent("databasing.StartMessageListening", func() { StartMessageListening(db) })

	}

}

var onClose func()

func End() {
	Events.FuncEvent("Databasing.End", func() {
		if onClose != nil {
			onClose()
		}
		close(queries)
		close(actions)
	})
}

func StartMessageListening(db *sql.DB) {
	for {
		select {
		case request := <-queries:
			if request == nil {
				return
			}
			go request.execute()
		case request := <-actions:
			if request == nil {
				return
			}
			go request.execute()
		}
	}
}

func SetupServer() {
	//LoadAllMembers()
	RequestChannel("AllNames")

}

func Close() {
	close(queries)
	close(actions)
}

func IsName(input string) bool {
	return reIsName.FindString(input) == input
}

func SanatizeDatabaseInput(input string) string {
	return reSanatizeDatabase.ReplaceAllStringFunc(input, func(match string) string {
		return "\\" + match
	})
}

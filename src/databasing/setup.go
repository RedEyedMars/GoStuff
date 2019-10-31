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

var ResourceRequests chan *DBResourceResponse
var ResourcesRequests chan *DBResourcesResponse
var ChatMsgRequests chan *DBChatMsgResponse
var MemberRequests chan *DBMemberResponse
var MemberNamesRequests chan *DBMemberResponse
var ChannelRequests chan *DBChannelResponse
var ChannelNamesRequests chan *DBClientChannelResponse
var ActionRequests chan *DBActionResponse

var LoadedResources map[string]*Resource
var Channels map[string]*Channel

var MembersByName map[string]*Member

var reSanatizeDatabase *regexp.Regexp
var reIsName *regexp.Regexp

var adminCommands map[string]Events.Event
var adminArgs []string

type DBActionResponse struct {
	Exec       func() (sql.Result, error)
	Successful chan bool
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
		adminCommands["addMember"] = &Events.Function{Name: "Admin!AddMember", Function: makeAdminFunc(1,
			func(args ...string) { RequestMemberAction("Add", NewMember(), args[0]) })}
		adminCommands["addMemberFull"] = &Events.Function{Name: "Admin!AddMember_Full", Function: makeAdminFunc(2,
			func(args ...string) { RequestMemberAction("Add", NewMemberFull(args[0]), args[1]) })}
		adminCommands["removeMember"] = &Events.Function{Name: "Admin!RemoveMember", Function: makeAdminFunc(1,
			func(args ...string) { RequestMemberAction("Remove", MembersByName[args[0]]) })}
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

	ResourceRequests = make(chan *DBResourceResponse, 16)
	ResourcesRequests = make(chan *DBResourcesResponse, 16)
	ChatMsgRequests = make(chan *DBChatMsgResponse, 16)
	MemberRequests = make(chan *DBMemberResponse, 16)
	MemberNamesRequests = make(chan *DBMemberResponse, 16)
	ChannelRequests = make(chan *DBChannelResponse, 16)
	ChannelNamesRequests = make(chan *DBClientChannelResponse, 16)
	ActionRequests = make(chan *DBActionResponse, 32)

	LoadedResources = make(map[string]*Resource)
	Channels = make(map[string]*Channel)
	MembersByName = make(map[string]*Member)

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

	dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbEndpoint, dbName)

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
		close(ResourceRequests)
		close(ResourcesRequests)
		close(ChatMsgRequests)
		close(MemberRequests)
		close(MemberNamesRequests)
		close(ChannelRequests)
		close(ChannelNamesRequests)
		close(ActionRequests)
	})
}

func StartMessageListening(db *sql.DB) {
	for {
		select {
		case request := <-ResourceRequests:
			if request == nil {
				return
			}
			row := request.QueryRow()
			Events.GoFuncEvent("databasing.Resources.Parse", func() { request.Parse(row) })
		case request := <-ResourcesRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ResourceRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.Resources.Parse", func() { request.Parse(rows) })
			}
		case request := <-ChatMsgRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChatMsgRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.chatmgs.Parse", func() { request.Parse(rows) })
			}
		case request := <-MemberRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.MemberRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.members.Parse", func() { request.Parse(rows) })
			}
		case request := <-MemberNamesRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.MemberNamesRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.members.ParseNames", func() { request.ParseNames(rows) })
			}
		case request := <-ChannelRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChannelRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.channels.ParseNew", func() { request.ParseNew(rows) })
			}
		case request := <-ChannelNamesRequests:
			if request == nil {
				return
			}
			if rows, err := request.Query(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChannelRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.channels.ParseNames", func() { request.ParseNames(rows) })
			}
		case request := <-ActionRequests:
			if request == nil {
				return
			}
			if result, err := request.Exec(); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ActionRequest.Exec"}
			} else {
				Events.GoFuncEvent("databasing.setup.Success", func() { request.ParseSuccess(result) })
			}

		}

	}
}

func SetupServer() {
	//LoadAllMembers()
	RequestChannel("AllNames")

}

func Close() {
	close(ResourceRequests)
	close(ResourcesRequests)
	close(ChatMsgRequests)
	close(MemberRequests)
	close(MemberNamesRequests)
	close(ChannelRequests)
	close(ChannelNamesRequests)
}

func IsName(input string) bool {
	return reIsName.FindString(input) == input
}

func SanatizeDatabaseInput(input string) string {
	return reSanatizeDatabase.ReplaceAllStringFunc(input, func(match string) string {
		return "\\" + match
	})
}

func (as *DBActionResponse) ParseSuccess(result sql.Result) {
	if _, err := result.RowsAffected(); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.setup.ParseSuccess"}
	}

	as.Successful <- true
	close(as.Successful)
}

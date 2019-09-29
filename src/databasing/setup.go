package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"fmt"
	"log"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

var dbQueries map[string]string
var dbQueryArgumentLength map[string]int

var ResourceRequests chan *DBResourceResponse
var ResourcesRequests chan *DBResourceResponse
var ChatMsgRequests chan *DBChatMsgResponse
var MemberRequests chan *DBMemberResponse
var MemberNamesRequests chan *DBMemberResponse
var ChannelRequests chan *DBChannelResponse
var ChannelNamesRequests chan *DBChannelResponse

var LoadedResources map[string]*Resource
var Channels map[string]*Channel

var MembersByName map[string]*Member
var MembersByIp map[string]*Member

var reSanatizeDatabase *regexp.Regexp
var reIsName *regexp.Regexp

func Setup() {
	dbQueries = make(map[string]string)
	dbQueryArgumentLength = make(map[string]int)

	ResourceRequests = make(chan *DBResourceResponse, 16)
	ResourcesRequests = make(chan *DBResourceResponse, 16)
	ChatMsgRequests = make(chan *DBChatMsgResponse, 16)
	MemberRequests = make(chan *DBMemberResponse, 16)
	MemberNamesRequests = make(chan *DBMemberResponse, 16)
	ChannelRequests = make(chan *DBChannelResponse, 16)
	ChannelNamesRequests = make(chan *DBChannelResponse, 16)

	LoadedResources = make(map[string]*Resource)
	Channels = make(map[string]*Channel)
	MembersByName = make(map[string]*Member)
	MembersByIp = make(map[string]*Member)

	Events.FuncEvent("databasing.SetupResources", SetupResources)
	Events.FuncEvent("databasing.SetupChatMsgs", SetupChatMsgs)
	Events.FuncEvent("databasing.SetupMembers", SetupMembers)
	Events.FuncEvent("databasing.SetupChannels", SetupChannels)

	reSanatizeDatabase = regexp.MustCompile(`(\n, \r, \, ', ")`)
	reIsName = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]*`)
}
func defineQuery(name string, query string, argLength int) {
	dbQueries[name] = query
	dbQueryArgumentLength[name] = argLength
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
		Events.FuncEvent("databasing.StartMessageListening", func() { StartMessageListening(db) })

	}

}

func End() {
	Events.FuncEvent("Databasing.End", func() {
		close(ResourceRequests)
		close(ResourcesRequests)
		close(ChatMsgRequests)
		close(MemberRequests)
		close(MemberNamesRequests)
		close(ChannelRequests)
		close(ChannelNamesRequests)
	})
}

func StartMessageListening(db *sql.DB) {
	for {
		select {
		case request := <-ResourceRequests:
			row := db.QueryRow(request.Query())
			Events.GoFuncEvent("databasing.Resources.Parse", func() { request.Parse(row) })
		case request := <-ResourcesRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ResourceRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.Resources.Parse", func() { request.ParseAll(rows) })
			}
		case request := <-ChatMsgRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChatMsgRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.chatmgs.Parse", func() { request.Parse(rows) })
			}
		case request := <-MemberRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.MemberRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.members.Parse", func() { request.Parse(rows) })
			}
		case request := <-MemberNamesRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.MemberRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.members.ParseNames", func() { request.ParseNames(rows) })
			}
		case request := <-ChannelRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChannelRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.channels.ParseNew", func() { request.ParseNew(rows) })
			}
		case request := <-ChannelNamesRequests:
			if rows, err := db.Query(request.Query()); err != nil {
				Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartMessageListening.ChannelRequest.Query"}
			} else {
				Events.GoFuncEvent("databasing.channels.ParseNames", func() { request.ParseNames(rows) })
			}
		default:
			Logger.Verbose <- Logger.Msg{"Closing database..."}
			db.Close()
			return
		}
	}
}

func SetupServer() {
	//LoadAllMembers()
	RequestChannel("AllNames")

}

func SetupMember(ip string) *Member {
	member, present := MembersByIp[ip]
	if !present {
		member, present = <-RequestMember("ByIp", ip)
		member = NewMember(ip)
	}
	for channel := range RequestChannelsByName("Members", member.Name) {
		channel.Members[member.Name] = member
	}
	return member
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

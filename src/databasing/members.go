package databasing

import (
	"Events"
	"Logger"
	"database/sql"
	"fmt"
	"math/rand"
)

/**
client_names
+-------+--------------+------+-----+---------+-------+
| Field | Type         | Null | Key | Default | Extra |
+-------+--------------+------+-----+---------+-------+
| name  | varchar(255) | NO   | PRI | NULL    |       |
| pwd   | varchar(255) | NO   | PRI | NULL    |       |
+-------+--------------+------+-----+---------+-------+

channels_names
+--------------+------------------+------+-----+---------+----------------+
| Field        | Type             | Null | Key | Default | Extra          |
+--------------+------------------+------+-----+---------+----------------+
| channel_name | varchar(255)     | NO   |     | NULL    |                |
| member_name  | varchar(255)     | YES  |     | NULL    |                |
| id           | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
+--------------+------------------+------+-----+---------+----------------+

**/

func LoadAllMembers() {

	for member := range RequestMember("All") {
		Events.FuncEvent("databasing.members.AddMemberToMap", func() { AddMemberToMaps(member) })
	}
}

type Member struct {
	Name string
}

type DBMemberResponse struct {
	Query   func() (*sql.Rows, error)
	Members chan *Member
}

func NewMember() *Member {
	member := &Member{
		Name: fmt.Sprintf("%s%s%s", Adverbs[rand.Intn(len(Adverbs))],
			Adjectives[rand.Intn(len(Adjectives))],
			Nouns[rand.Intn(len(Nouns))])}
	Events.FuncEvent("databasing.members.AddMemberToMap", func() { AddMemberToMaps(member) })
	return member
}
func NewMemberFull(name string) *Member {
	member := &Member{
		Name: name}
	Events.FuncEvent("databasing.members.AddMemberToMap", func() { AddMemberToMaps(member) })
	return member
}
func AddMemberToMaps(member *Member) {
	Logger.Verbose <- Logger.Msg{"Add Member: " + member.Name}
	MembersByName[member.Name] = member
}

func NewMemberResponse(name string, arg ...interface{}) *DBMemberResponse {
	return NewMemberResponseArr(name, arg)
}
func NewMemberResponseArr(name string, args []interface{}) *DBMemberResponse {
	return &DBMemberResponse{
		Query:   func() (*sql.Rows, error) { return dbQueries["Members_"+name].Query(args...) },
		Members: make(chan *Member, 1)}
}
func NewMemberActionArr(name string, args []interface{}) *DBActionResponse {
	return &DBActionResponse{
		Exec:       func() (sql.Result, error) { return dbQueries["Members_"+name].Exec(args...) },
		Successful: make(chan bool, 1)}
}

func SetupMembers(db *sql.DB) {
	defineQuery(db, "Members_All", `SELECT name FROM client_names ;`)

	defineQuery(db, "Members_ByName", `SELECT name FROM client_names WHERE name=? ;`)
	defineQuery(db, "Members_ByPwd", `SELECT name FROM client_names WHERE pwd=? ;`)

	defineQuery(db, "Members_Add", `INSERT INTO client_names VALUES (?,?);`)
	defineQuery(db, "Members_Remove", `DELETE FROM client_names WHERE pwd = ?;`)
}

func RequestMember(name string, args ...interface{}) <-chan *Member {
	request := NewMemberResponseArr(name, args)
	MemberRequests <- request
	return request.Members
}
func RequestMembersByName(name string, args ...interface{}) <-chan *Member {
	request := NewMemberResponseArr(name, args)
	MemberNamesRequests <- request
	return request.Members
}
func RequestMemberAction(name string, member *Member, args ...interface{}) <-chan bool {
	var request *DBActionResponse
	switch len(args) {
	case 0:
		request = NewMemberActionArr(name, []interface{}{member.Name})
	case 1:
		request = NewMemberActionArr(name, []interface{}{member.Name, args[0]})
	case 2:
		request = NewMemberActionArr(name, []interface{}{member.Name, args[0], args[1]})
	}
	ActionRequests <- request
	return request.Successful
}
func (ms *DBMemberResponse) Parse(rows *sql.Rows) {
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.Parse"}
		}

		ms.Members <- &Member{Name: name}
	}
	close(ms.Members)
}
func (ms *DBMemberResponse) ParseNames(rows *sql.Rows) {
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.ParseNames"}
		}

		ms.Members <- MembersByName[name]
	}
	close(ms.Members)
}

package databasing

import (
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
| ip    | varchar(15)  | NO   | PRI | NULL    |       |
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
		MembersByName[member.Name] = member
		MembersByIp[member.IP] = member
	}
}

type Member struct {
	Name string
	IP   string
}

type DBMemberResponse struct {
	Query   func() (*sql.Rows, error)
	Members chan *Member
}

func NewMember(ip string) *Member {
	member := &Member{
		Name: fmt.Sprintf("%s%s%s", Adverbs[rand.Intn(len(Adverbs))],
			Adjectives[rand.Intn(len(Adjectives))],
			Nouns[rand.Intn(len(Nouns))]),
		IP: ip}
	AddMemberToMaps(member)
	return member
}
func AddMemberToMaps(member *Member) {
	Logger.Verbose <- Logger.Msg{"Add Member: " + member.Name + "; " + member.IP}
	MembersByName[member.Name] = member
	MembersByIp[member.IP] = member
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
	defineQuery(db, "Members_All", `SELECT name,ip FROM client_names ;`)

	defineQuery(db, "Members_ByName", `SELECT name,ip FROM client_names WHERE name=? ;`)
	defineQuery(db, "Members_ByIp", `SELECT name,ip FROM client_names WHERE ip=? ;`)

	defineQuery(db, "Members_Add", `INSERT INTO client_names VALUES (?,?);`)
	defineQuery(db, "Members_Remove", `DELETE FROM client_names WHERE name = ? and ip = ?;`)
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
func RequestMemberAction(name string, member *Member) <-chan bool {
	request := NewMemberActionArr(name, []interface{}{member.Name, member.IP})
	ActionRequests <- request
	return request.Successful
}
func (ms *DBMemberResponse) Parse(rows *sql.Rows) {
	for rows.Next() {
		var (
			name string
			ip   string
		)
		if err := rows.Scan(&name, &ip); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.Parse"}
		}

		ms.Members <- &Member{name, ip}
	}
	close(ms.Members)
}
func (ms *DBMemberResponse) ParseNames(rows *sql.Rows) {
	for rows.Next() {
		var (
			name string
			ip   string
		)
		if err := rows.Scan(&name, &ip); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.ParseNames"}
		}

		ms.Members <- MembersByName[name]
	}
	close(ms.Members)
}

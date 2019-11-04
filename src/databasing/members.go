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

var Members map[string]*Member

func LoadAllMembers() {

	for member := range RequestMember("All") {
		Events.FuncEvent("databasing.members.AddMemberToMap", func() { AddMemberToMaps(member) })
	}
}

type Member struct {
	Name string
}

type DBMemberResponse struct {
	chl       chan *Member
	assembler func(*sql.Rows) *Member
}

func (mr *DBMemberResponse) send(result *sql.Rows) {
	mr.chl <- mr.assembler(result)
}
func (mr *DBMemberResponse) close() {
	close(mr.chl)
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
	Members[member.Name] = member
}

func SetupMembers(db *sql.DB) {
	defineQuery(db, "Members_All", `SELECT name FROM client_names ;`)

	defineQuery(db, "Members_ByName", `SELECT name FROM client_names WHERE name=? ;`)
	defineQuery(db, "Members_ByPwd", `SELECT name FROM client_names WHERE pwd=? ;`)

	defineQuery(db, "Members_Add", `INSERT INTO client_names VALUES (?,?);`)
	defineQuery(db, "Members_Remove", `DELETE FROM client_names WHERE name = ?;`)
}

func RequestMember(name string, args ...interface{}) <-chan *Member {
	response := make(chan *Member, 1)
	queries <- &DBQueryResponse{
		query: "Members_" + name,
		args:  args,
		sender: &DBMemberResponse{
			chl:       response,
			assembler: parseMember,
		},
	}
	return response
}
func RequestMembersByName(name string, args ...interface{}) <-chan *Member {
	response := make(chan *Member, 1)
	queries <- &DBQueryResponse{
		query: "Members_" + name,
		args:  args,
		sender: &DBMemberResponse{
			chl:       response,
			assembler: parseMemberByName,
		},
	}
	return response
}
func parseMember(rows *sql.Rows) *Member {
	var name string
	if err := rows.Scan(&name); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.Parse"}
	}
	return NewMemberFull(name)
}
func parseMemberByName(rows *sql.Rows) *Member {
	var name string
	if err := rows.Scan(&name); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.ParseNames"}
	}

	return Members[name]
}

package databasing

import (
	"Logger"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
)

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
	Query   func() string
	Members chan *Member
}

func NewMember(ip string) *Member {
	return &Member{
		Name: fmt.Sprintf("%s%s%s", Adverbs[rand.Intn(len(Adverbs))],
			Adjectives[rand.Intn(len(Adjectives))],
			Nouns[rand.Intn(len(Nouns))]),
		IP: ip}
}

func NewMemberResponse(name string, arg ...string) *DBMemberResponse {
	return NewMemberResponseArr(name, arg)
}
func NewMemberResponseArr(name string, arg []string) *DBMemberResponse {
	name = "Members_" + name
	switch dbQueryArgumentLength[name] {
	case 0:
		return &DBMemberResponse{
			Query:   func() string { return dbQueries[name] },
			Members: make(chan *Member, 1)}
	case 1:
		return &DBMemberResponse{
			Query:   func() string { return fmt.Sprintf(dbQueries[name], arg[0]) },
			Members: make(chan *Member, 1)}
	case 2:
		return &DBMemberResponse{
			Query:   func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1]) },
			Members: make(chan *Member, 1)}
	case 3:
		return &DBMemberResponse{
			Query:   func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1], arg[2]) },
			Members: make(chan *Member, 1)}
	default:
		Logger.Error <- Logger.ErrMsg{Err: errors.New(name + "has to many arguments:" + strconv.Itoa(dbQueryArgumentLength[name])), Status: "databasing.members.NewMemberResponse"}
		return nil
	}
}

func SetupMembers() {
	defineQuery("Members_All", `SELECT name,ip FROM client_names ;`, 0)

	defineQuery("Members_ByName", `SELECT name,ip FROM client_names WHERE name='%s' ;`, 1)
	defineQuery("Members_ByIp", `SELECT name,ip FROM client_names WHERE ip='%s' ;`, 1)
}
func RequestMember(name string, args ...string) <-chan *Member {
	request := NewMemberResponseArr(name, args)
	MemberRequests <- request
	return request.Members
}
func RequestMembersByName(name string, args ...string) <-chan *Member {
	request := NewMemberResponseArr(name, args)
	MemberNamesRequests <- request
	return request.Members
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
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.members.Parse"}
		}

		ms.Members <- MembersByName[name]
	}
	close(ms.Members)
}

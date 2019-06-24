package databasing

import (
	"Logger"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Resource struct {
	Abv    string
	Name   string
	Source string
}

type DBResourceResponse struct {
	Query     func() string
	Resources chan *Resource
}

func NewResourceResponse(name string, arg ...string) *DBResourceResponse {
	return NewResourceResponseArr(name, arg)
}
func NewResourceResponseArr(name string, arg []string) *DBResourceResponse {
	name = "Resource_" + name
	switch dbQueryArgumentLength[name] {
	case 0:
		return &DBResourceResponse{
			Query:     func() string { return dbQueries[name] },
			Resources: make(chan *Resource, 1)}
	case 1:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0]) },
			Resources: make(chan *Resource, 1)}
	case 2:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1]) },
			Resources: make(chan *Resource, 1)}
	case 3:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1], arg[2]) },
			Resources: make(chan *Resource, 1)}
	default:
		Logger.Error <- Logger.ErrMsg{Err: errors.New(name + "has to many arguments:" + strconv.Itoa(dbQueryArgumentLength[name])), Status: "resources.NewResourceResponse"}
		return nil
	}
}

func SetupResources() {
	defineQuery("Resources_ByName", `SELECT name, source, abv FROM resources WHERE name='%s' ;`, 1)
	defineQuery("Resources_ByAbv", `SELECT name, source, abv FROM resources WHERE abv='%s' ;`, 1)
	defineQuery("Resources_BySource", `SELECT name, source, abv FROM resources WHERE source='%s' ;`, 1)
}
func RequestResource(name string, args ...string) <-chan *Resource {
	request := NewResourceResponseArr(name, args)
	ResourceRequests <- request
	return request.Resources
}
func (r *DBResourceResponse) ParseAll(rows *sql.Rows) {
	for rows.Next() {
		var (
			name   string
			source string
			abv    string
		)
		if err := rows.Scan(&name, &source, &abv); err != nil {
			Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.resources.Parse"}
		}
		r.Resources <- &Resource{name, source, abv}
	}
	close(r.Resources)
}
func (r *DBResourceResponse) Parse(row *sql.Row) {
	var (
		name   string
		source string
		abv    string
	)
	if err := row.Scan(&name, &source, &abv); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.resources.Parse"}
	}
	r.Resources <- &Resource{name, source, abv}
	close(r.Resources)
}

package databasing

import (
	"Logger"
	"database/sql"
)

type Resource struct {
	Abv    string
	Name   string
	Source string
}

type DBResourceResponse struct {
	QueryRow  func() *sql.Row
	Resources chan *Resource
}

type DBResourcesResponse struct {
	Query     func() (*sql.Rows, error)
	Resources chan *Resource
}

func NewResourceResponse(name string, arg ...string) *DBResourceResponse {
	return NewResourceResponseArr(name, arg)
}
func NewResourceResponseArr(name string, args []string) *DBResourceResponse {
	name = "Resource_" + name
	return &DBResourceResponse{
		QueryRow:  func() *sql.Row { return dbQueries["Resource_"+name].QueryRow(args) },
		Resources: make(chan *Resource, 1)}
}
func NewResourcesResponseArr(name string, args []string) *DBResourcesResponse {
	name = "Resource_" + name
	return &DBResourcesResponse{
		Query:     func() (*sql.Rows, error) { return dbQueries["Resource_"+name].Query(args) },
		Resources: make(chan *Resource, 1)}
}

func SetupResources(db *sql.DB) {
	defineQuery(db, "Resources_ByName", `SELECT name, source, abv FROM resources WHERE name=? ;`)
	defineQuery(db, "Resources_ByAbv", `SELECT name, source, abv FROM resources WHERE abv=? ;`)
	defineQuery(db, "Resources_BySource", `SELECT name, source, abv FROM resources WHERE source=? ;`)
}
func RequestResource(name string, args ...string) <-chan *Resource {
	request := NewResourceResponseArr(name, args)
	ResourceRequests <- request
	return request.Resources
}
func (r *DBResourcesResponse) Parse(rows *sql.Rows) {
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

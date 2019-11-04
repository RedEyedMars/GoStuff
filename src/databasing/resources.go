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
	chl       chan *Resource
	assembler func(*sql.Rows) *Resource
}

func (mr *DBResourceResponse) send(result *sql.Rows) {
	mr.chl <- mr.assembler(result)
}
func (mr *DBResourceResponse) close() {
	close(mr.chl)
}

func SetupResources(db *sql.DB) {
	defineQuery(db, "Resources_ByName", `SELECT name, source, abv FROM resources WHERE name=? ;`)
	defineQuery(db, "Resources_ByAbv", `SELECT name, source, abv FROM resources WHERE abv=? ;`)
	defineQuery(db, "Resources_BySource", `SELECT name, source, abv FROM resources WHERE source=? ;`)
}
func RequestResource(name string, args ...interface{}) <-chan *Resource {
	response := make(chan *Resource, 1)
	queries <- &DBQueryResponse{
		query: "Resources_" + name,
		args:  args,
		sender: &DBResourceResponse{
			chl:       response,
			assembler: parseResources,
		},
	}
	return response
}
func parseResources(row *sql.Rows) *Resource {
	var (
		name   string
		source string
		abv    string
	)
	if err := row.Scan(&name, &source, &abv); err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "databasing.resources.Parse"}
	}
	return &Resource{name, source, abv}
}

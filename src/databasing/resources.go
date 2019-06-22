package databasing

import (
	"Logger"
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
	Resources chan Resource
}

func NewResourceResponse(name string, arg ...string) *DBResourceResponse {
	name = "Resource_" + name
	switch dbQueryArgumentLength[name] {
	case 0:
		return &DBResourceResponse{
			Query:     func() string { return dbQueries[name] },
			Resources: make(chan Resource, 1)}
	case 1:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0]) },
			Resources: make(chan Resource, 1)}
	case 2:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1]) },
			Resources: make(chan Resource, 1)}
	case 3:
		return &DBResourceResponse{
			Query:     func() string { return fmt.Sprintf(dbQueries[name], arg[0], arg[1], arg[2]) },
			Resources: make(chan Resource, 1)}
	default:
		Logger.Error <- Logger.ErrMsg{errors.New(name + "has to many arguments:" + strconv.Itoa(dbQueryArgumentLength[name])), "resources.NewResourceResponse"}
		return nil
	}
}

func SetupResources() {
	defineQuery("Resources_ByName", `SELECT name, source, abv FROM resources WHERE name='%s' ;`, 1)
	defineQuery("Resources_ByAbv", `SELECT name, source, abv FROM resources WHERE abv='%s' ;`, 1)
	defineQuery("Resources_BySource", `SELECT name, source, abv FROM resources WHERE source='%s' ;`, 1)
}

func defineQuery(name string, query string, argLength int) {
	dbQueries[name] = query
	dbQueryArgumentLength[name] = argLength
}

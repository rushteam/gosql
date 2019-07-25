package pool

import (
	"database/sql"
	"errors"
	"math/rand"
)

type Cluster struct {
	dbType   string
	settings map[string]map[string][]string
	pool     map[string]*sql.DB
}

func (c Cluster) Get(name, node string) (*sql.DB, error) {
	var conf string
	if setting, ok := c.settings[name]; ok {
		if _, ok := setting["master"]; !ok {
			return nil, errors.New("master config is undefined")
		}
		if _, ok := setting[node]; !ok {
			setting[node] = setting["master"]
		}
		idx := rand.Intn(len(setting[node]))
		conf = setting[node][idx]
	}
	if conf == "" {
		return nil, errors.New("db config should be not empty")
	}
	// conf = "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true"
	if db, ok := c.pool[conf]; ok {
		return db, nil
	}
	db, err := sql.Open(c.dbType, conf)
	if err != nil {
		return nil, err
	}
	c.pool[conf] = db
	return c.pool[conf], nil
}

//Init ..
func Init(dbType string, settings map[string]map[string][]string) *Cluster {
	c := &Cluster{}
	c.dbType = dbType
	c.settings = settings
	c.pool = make(map[string]*sql.DB, 0)
	return c
}

//Start ..
func Start() {
	go func() {

	}()
}

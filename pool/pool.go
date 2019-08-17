package pool

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"
)

var defaultCluster *Cluster

//Cluster ..
type Cluster struct {
	dbType   string
	settings map[string]map[string][]string
	pool     map[string]*sql.DB
}
func(c Cluster) open() {
	var dsn string
	if setting, ok := c.settings[name]; ok {
		if _, ok := setting["master"]; !ok {
			return nil, errors.New("master dsn is undefined")
		}
		if _, ok := setting[node]; !ok {
			setting[node] = setting["master"]
		}
		idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(setting[node]))
		dsn = setting[node][idx]
	}
	// if dsn == "" {
		// return nil, errors.New("db DSN should be not empty")
	// }
	// conf = "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true"
	connFn := func() error {
		sess, err := sql.Open(c.dbType, dsn)
		if err == nil {
			sess.SetConnMaxLifetime(db.DefaultSettings.ConnMaxLifetime())
			sess.SetMaxIdleConns(db.DefaultSettings.MaxIdleConns())
			sess.SetMaxOpenConns(db.DefaultSettings.MaxOpenConns())
			return d.BaseDatabase.BindSession(sess)
		}
		return err
	}
}
//Open ..
func (c Cluster) Open(name, node string) (*sql.DB, error) {
	var dsn string
	if setting, ok := c.settings[name]; ok {
		if _, ok := setting["master"]; !ok {
			return nil, errors.New("master dsn is undefined")
		}
		if _, ok := setting[node]; !ok {
			setting[node] = setting["master"]
		}
		idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(setting[node]))
		dsn = setting[node][idx]
	}
	if dsn == "" {
		return nil, errors.New("db DSN should be not empty")
	}
	// conf = "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true"
	if db, ok := c.pool[dsn]; ok {
		return db, nil
	}
	db, err := sql.Open(c.dbType, dsn)
	if err != nil {
		return nil, err
	}
	c.pool[dsn] = db
	return c.pool[dsn], nil
}

//Init ..
func Init(dbType string, settings map[string]map[string][]string) *Cluster {
	c := &Cluster{}
	c.dbType = dbType
	c.settings = settings
	c.pool = make(map[string]*sql.DB, 0)
	defaultCluster = c
	return c
}

// //Open Session
// func (s *Session) Open(clusterName, clusterNode string) (*Executor error) {
// 	var err error
// 	s.exec, err = defaultCluster.Open(clusterName, clusterNode)
// 	return s.exec, err
// }

// //Model Session
// func (s *Session) Model(dst interface{}) *ORM {
// 	o := &ORM{}
// 	o.Ctor(dst, s)
// 	return o
// }

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

//Opts ..
type Opts func(db *sql.DB) *sql.DB

//PoolCluster ..
type PoolCluster struct {
	dbType   string
	settings map[string][]string
	pool     map[string]*sql.DB
	idx      uint64
	opts     []Opts
	dbOpt    struct {
		connMaxLifetime time.Duration
		MaxIdleConns    int
		MaxOpenConns    int
	}
}

//open ..
func (c *PoolCluster) open(dbType string, dsn string, opts ...Opts) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("db: DSN should be not empty")
	}
	//如果已经存在
	if db, ok := c.pool[dsn]; ok {
		return db, nil
	}
	db, err := sql.Open(c.dbType, dsn)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		db = opt(db)
	}
	c.pool[dsn] = db
	return c.pool[dsn], nil
}

//Open ..
func (c *PoolCluster) Open(dsn string) (*sql.DB, error) {
	opt := func(db *sql.DB) *sql.DB {
		db.SetConnMaxLifetime(c.dbOpt.connMaxLifetime)
		db.SetMaxIdleConns(c.dbOpt.MaxIdleConns)
		db.SetMaxOpenConns(c.dbOpt.MaxOpenConns)
		return db
	}
	return c.open(c.dbType, dsn, opt)
}

//Master ..
func (c *PoolCluster) Master() (Executor, error) {
	name := "default"
	if setting, ok := c.settings[name]; ok {
		debugPrint("db: [master] %s\r\n", setting[0])
		return c.Open(setting[0])
	}
	return nil, nil
}

//Slave ..
func (c *PoolCluster) Slave() (Executor, error) {
	name := "default"
	if setting, ok := c.settings[name]; ok {
		var i int
		n := len(setting) - 1
		//idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(setting[node]))
		v := atomic.AddUint64(&c.idx, 1)
		if n > 0 {
			i = int(v)%(n) + 1
		}
		debugPrint("db: [slave#%d] %s\r\n", i, setting[i])
		return c.Open(setting[0])
	}
	return nil, nil
}

//Begin ..
func (c *PoolCluster) Begin() (*sql.Tx, error) {
	ex, err := c.Master()
	if err != nil {
		return nil, err
	}
	if db, ok := ex.(Db); ok {
		return db.Begin()
	}
	return nil, errors.New("db: not Db type")
}

func debugPrint(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}

//InitPool ..
func InitPool(dbType string, settings map[string][]string, opts ...Opts) *PoolCluster {
	c := &PoolCluster{}
	c.idx = 0
	c.dbType = dbType
	c.settings = settings
	c.pool = make(map[string]*sql.DB, len(settings))
	c.opts = opts
	commonSession = &Session{cluster: c, master: false}
	return c
}

//SetConnMaxLifetime ..
func SetConnMaxLifetime(d time.Duration) Opts {
	return func(db *sql.DB) *sql.DB {
		db.SetConnMaxLifetime(d)
		return db
	}
}

//SetMaxIdleConns ..
func SetMaxIdleConns(n int) Opts {
	return func(db *sql.DB) *sql.DB {
		db.SetMaxIdleConns(n)
		return db
	}
}

//SetMaxOpenConns ..
func SetMaxOpenConns(n int) Opts {
	return func(db *sql.DB) *sql.DB {
		db.SetMaxOpenConns(n)
		return db
	}
}

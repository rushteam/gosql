package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
)

//Cluster ..
type Cluster interface {
	Open(name, node string) (*sql.DB, error)
	Master() (Executor, error)
	Slave() (Executor, error)
}

//PoolCluster ..
type PoolCluster struct {
	dbType   string
	settings map[string][]string
	pool     map[string]*sql.DB
	idx      uint64
}

//Open ..
func (c *PoolCluster) Open(dbType string, dsn string) (*sql.DB, error) {
	// dsn = "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true"
	if dsn == "" {
		return nil, errors.New("db DSN should be not empty")
	}
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

//Master ..
func (c *PoolCluster) Master() (Executor, error) {
	name := "default"
	if setting, ok := c.settings[name]; ok {
		debugPrint("db: [master] %s\r\n", setting[0])
		return c.Open(c.dbType, setting[0])
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
		return c.Open(c.dbType, setting[i])
	}
	return nil, nil
}
func debugPrint(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}

//InitPool ..
func InitPool(dbType string, settings map[string][]string) *PoolCluster {
	c := &PoolCluster{}
	c.idx = 0
	c.dbType = dbType
	c.settings = settings
	c.pool = make(map[string]*sql.DB, len(settings))
	return c
}

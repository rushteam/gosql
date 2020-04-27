package gosql

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"
)

type dbEngine struct {
	Db              *sql.DB
	Dsn             string
	Driver          string
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

//Connect real open a db
func (d *dbEngine) Connect() (*sql.DB, error) {
	if d.Db == nil {
		db, err := sql.Open(d.Driver, d.Dsn)
		if err != nil {
			return nil, err
		}
		d.Db = db
	}
	return d.Db, nil
}

//PoolCluster ..
type PoolCluster struct {
	vs      uint64
	pools   []*dbEngine
	session *Session
}

//PoolClusterOpts ..
type PoolClusterOpts func(p *PoolCluster) *PoolCluster

//Session ..
func (c *PoolCluster) Session() (*Session, error) {
	return c.SessionContext(context.TODO())
}

//SessionContext ..
func (c *PoolCluster) SessionContext(ctx context.Context) (*Session, error) {
	if c.session == nil {
		v := atomic.AddUint64(&(c.vs), 1)
		c.session = &Session{ctx: ctx, cluster: c, v: v}
	}
	return c.session, nil
}

//Master select db to master
func (c *PoolCluster) Master() (Executor, error) {
	if len(c.pools) > 0 {
		dbx := c.pools[0]
		debugPrint("db: [master] %s", dbx.Dsn)
		return dbx.Connect()
	}
	return nil, errors.New("not found master db")
}

//Slave select db to slave
func (c *PoolCluster) Slave(v uint64) (Executor, error) {
	var i int
	n := len(c.pools) - 1
	if n > 0 {
		i = int(v)%(n) + 1
	}
	if len(c.pools) > 0 {
		dbx := c.pools[i]
		debugPrint("db: [slave#%d] %s", dbx.Dsn)
		return dbx.Connect()
	}
	return nil, errors.New("not found slave db")
}

//Begin a transaction
func (c *PoolCluster) Begin() (*Session, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Begin", s.v)
	executor, err := s.Executor(true)
	if err != nil {
		return nil, err
	}
	executor, err = executor.(DB).Begin()
	if err != nil {
		return nil, err
	}
	s.executor = executor
	return s, nil
}

//Fetch fetch record to model
func (c *PoolCluster) Fetch(dst interface{}, opts ...Option) error {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Fetch", s.v)
	return s.Fetch(dst, opts...)
}

//FetchAll fetch records to models
func (c *PoolCluster) FetchAll(dst interface{}, opts ...Option) error {
	s, _ := c.Session()
	debugPrint("db: [session #%v] FetchAll", s.v)
	return s.FetchAll(dst, opts...)
}

//Update update from model
func (c *PoolCluster) Update(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Update", s.v)
	return s.Update(dst, opts...)
}

//Insert insert from model
func (c *PoolCluster) Insert(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Insert", s.v)
	return s.Insert(dst, opts...)
}

//Replace replace from model
func (c *PoolCluster) Replace(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Replace", s.v)
	return s.Replace(dst, opts...)
}

//Delete delete record
func (c *PoolCluster) Delete(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Delete", s.v)
	return s.Delete(dst, opts...)
}

//NewCluster start
func NewCluster(opts ...PoolClusterOpts) *PoolCluster {
	c := &PoolCluster{}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

//AddDb add a db
func AddDb(driver, dsn string) PoolClusterOpts {
	return func(p *PoolCluster) *PoolCluster {
		p.pools = append(p.pools, &dbEngine{
			Driver: driver,
			Dsn:    dsn,
		})
		return p
	}
}

//DbOpts ..
// type DbOpts func(db *sql.DB) *sql.DB
// //SetConnMaxLifetime ..
// func SetConnMaxLifetime(d time.Duration) DbOpts {
// 	return func(db *sql.DB) *sql.DB {
// 		db.SetConnMaxLifetime(d)
// 		return db
// 	}
// }

// //SetMaxIdleConns ..
// func SetMaxIdleConns(n int) DbOpts {
// 	return func(db *sql.DB) *sql.DB {
// 		db.SetMaxIdleConns(n)
// 		return db
// 	}
// }

// //SetMaxOpenConns ..
// func SetMaxOpenConns(n int) DbOpts {
// 	return func(db *sql.DB) *sql.DB {
// 		db.SetMaxOpenConns(n)
// 		return db
// 	}
// }
/*
//SetConnMaxLifetime ..
func SetConnMaxLifetime(d time.Duration) PoolClusterOpts {
	return func(p *PoolCluster) *PoolCluster {
		p.ConnMaxLifetime = d
		return p
	}
}

//SetMaxIdleConns ..
func SetMaxIdleConns(n int) DbOpts {
	return func(db *sql.DB) *sql.DB {
		p.MaxIdleConns = n
		return db
	}
}

//SetMaxOpenConns ..
func SetMaxOpenConns(n int) DbOpts {
	return func(db *sql.DB) *sql.DB {
		p.MaxOpenConns = n
		return db
	}
}
*/

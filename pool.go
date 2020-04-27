package godb

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"
)

//DbOpts ..
// type DbOpts func(db *sql.DB) *sql.DB

/*
//connect ..
func (c *PoolCluster) connect(dbType string, dsn string, opts ...DbOpts) (*sql.DB, error) {
	db, err := sql.Open(c.dbType, dsn)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		db = opt(db)
	}
	return db, nil
}

//Open ..
func (c *PoolCluster) Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("db: DSN should not be empty")
	}
	//如果已经存在
	if db, ok := c.pool[dsn]; ok {
		return db, nil
	}
	opt := func(db *sql.DB) *sql.DB {
		db.SetConnMaxLifetime(c.dbOpt.ConnMaxLifetime)
		db.SetMaxIdleConns(c.dbOpt.MaxIdleConns)
		db.SetMaxOpenConns(c.dbOpt.MaxOpenConns)
		return db
	}
	db, err := c.connect(c.dbType, dsn, opt)
	c.pool[dsn] = db
	return db, err
}

//Master ..
func (c *PoolCluster) Master() (Executor, error) {
	name := "default"
	if setting, ok := c.settings[name]; ok {
		debugPrint("db: [master] %s", setting[0])
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
		v := atomic.AddUint64(&c.idx, 1)
		if n > 0 {
			i = int(v)%(n) + 1
		}
		debugPrint("db: [slave#%d] %s", i, setting[i])
		return c.Open(setting[0])
	}
	return nil, nil
}

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


//InitPool ..
func InitPool(dbType string, settings map[string][]string, opts ...DbOpts) *PoolCluster {
	c := &PoolCluster{}
	c.idx = 0
	c.dbType = dbType
	c.settings = settings
	c.pool = make(map[string]*sql.DB, len(settings))
	c.opts = opts
	commonSession = NewSession(context.TODO(), c)
	return c
}
*/
type dbEngine struct {
	Db              *sql.DB
	Dsn             string
	Driver          string
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

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

//Master ..
func (c *PoolCluster) Master() (Executor, error) {
	if len(c.pools) > 0 {
		dbx := c.pools[0]
		debugPrint("db: [master] %s", dbx.Dsn)
		return dbx.Connect()
	}
	return nil, errors.New("not found master db")
}

//Slave ..
func (c *PoolCluster) Slave(v int) (Executor, error) {
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

//Begin ..
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

//Fetch ..
func (c *PoolCluster) Fetch(dst interface{}, opts ...Option) error {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Fetch", s.v)
	return s.Fetch(dst, opts...)
}

//FetchAll ..
func (c *PoolCluster) FetchAll(dst interface{}, opts ...Option) error {
	s, _ := c.Session()
	debugPrint("db: [session #%v] FetchAll", s.v)
	return s.FetchAll(dst, opts...)
}

//Update ..
func (c *PoolCluster) Update(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Update", s.v)
	return s.Update(dst, opts...)
}

//Insert ..
func (c *PoolCluster) Insert(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Insert", s.v)
	return s.Insert(dst, opts...)
}

//Replace ..
func (c *PoolCluster) Replace(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Replace", s.v)
	return s.Replace(dst, opts...)
}

//Delete ..
func (c *PoolCluster) Delete(dst interface{}, opts ...Option) (Result, error) {
	s, _ := c.Session()
	debugPrint("db: [session #%v] Delete", s.v)
	return s.Delete(dst, opts...)
}

//NewCluster ..
func NewCluster(opts ...PoolClusterOpts) *PoolCluster {
	c := &PoolCluster{}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

//AddDb ..
func AddDb(driver, dsn string) PoolClusterOpts {
	return func(p *PoolCluster) *PoolCluster {
		p.pools = append(p.pools, &dbEngine{
			Driver: driver,
			Dsn:    dsn,
		})
		return p
	}
}

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

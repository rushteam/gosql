package gosql

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"
	"unsafe"
)

//DbOption ..
type DbOption func(db *sql.DB) *sql.DB

type dbEngine struct {
	Db     *sql.DB
	Dsn    string
	Driver string
	Opts   []DbOption
}

//Connect real open a db
func (d *dbEngine) Connect() (*sql.DB, error) {
	if d.Db == nil {
		debugPrint("db: Connect(%s,%s)", d.Driver, d.Dsn)
		db, err := sql.Open(d.Driver, d.Dsn)
		if err != nil {
			return db, err
		}
		for _, opt := range d.Opts {
			db = opt(db)
		}
		d.Db = db
	}
	return d.Db, nil
}

//PoolCluster impl Cluster
type PoolCluster struct {
	vs      uint64
	pools   []*dbEngine
	session *Session
}

// PoolClusterOpts ..
type PoolClusterOpts func(p *PoolCluster) *PoolCluster

// NewSession ..
func (c *PoolCluster) NewSession() *Session {
	v := atomic.AddUint64(&(c.vs), 1)
	return &Session{cluster: c, v: v}
}

// Executor ..
func (c *PoolCluster) Executor(s *Session, master bool) (Executor, error) {
	if s == nil {
		s = c.NewSession()
	}
	var err error
	var executor Executor
	if master || s.forceMaster == true {
		executor, err = c.Master()
		s.setExecutor(executor)
	} else {
		executor, err = c.Slave(s.v)
		s.setExecutor(executor)
	}
	return executor, err
}

//Master select master db
func (c *PoolCluster) Master() (Executor, error) {
	if len(c.pools) == 0 {
		return nil, errors.New("not found db")
	}
	dbx := c.pools[0]
	debugPrint("db: [master] dsn %s", dbx.Dsn)
	return dbx.Connect()
}

//Slave select slave db
func (c *PoolCluster) Slave(v uint64) (Executor, error) {
	n := len(c.pools)
	if n == 0 {
		return nil, errors.New("not found db")
	}
	var i int
	if n > 1 {
		i = 1 + int(v)%(n-1)
	}
	dbx := c.pools[i]
	debugPrint("db: [slave#%d] %s", i, dbx.Dsn)
	return dbx.Connect()
}

//Begin a transaction
func (c *PoolCluster) Begin() (*Session, error) {
	s := c.NewSession()
	err := s.begin()
	return s, err
}

//QueryContext ..
func (c *PoolCluster) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := c.Executor(nil, false)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

//Query ..
func (c *PoolCluster) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := c.Executor(nil, false)
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

//QueryRowContext ..
func (c *PoolCluster) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := c.Executor(nil, false)
	if err != nil {
		row := &sql.Row{}
		rowErr := (*error)(unsafe.Pointer(row))
		*rowErr = err
		return row
	}
	return db.QueryRowContext(ctx, query, args...)
}

//QueryRow ..
func (c *PoolCluster) QueryRow(query string, args ...interface{}) *sql.Row {
	db, err := c.Executor(nil, false)
	if err != nil {
		row := &sql.Row{}
		rowErr := (*error)(unsafe.Pointer(row))
		*rowErr = err
		return row
	}
	return db.QueryRow(query, args...)
}

//ExecContext ..
func (c *PoolCluster) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

//Exec ..
func (c *PoolCluster) Exec(query string, args ...interface{}) (sql.Result, error) {
	db, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

//Fetch fetch record to model
func (c *PoolCluster) Fetch(dst interface{}, opts ...Option) error {
	s := c.NewSession()
	debugPrint("db: [session #%v] Fetch()", s.v)
	return s.Fetch(dst, opts...)
}

//FetchAll fetch records to models
func (c *PoolCluster) FetchAll(dst interface{}, opts ...Option) error {
	s := c.NewSession()
	debugPrint("db: [session #%v] FetchAll()", s.v)
	return s.FetchAll(dst, opts...)
}

//Update update from model
func (c *PoolCluster) Update(dst interface{}, opts ...Option) (Result, error) {
	s := c.NewSession()
	debugPrint("db: [session #%v] Update", s.v)
	return s.Update(dst, opts...)
}

//Insert insert from model
func (c *PoolCluster) Insert(dst interface{}, opts ...Option) (Result, error) {
	s := c.NewSession()
	debugPrint("db: [session #%v] Insert", s.v)
	return s.Insert(dst, opts...)
}

//Replace replace from model
func (c *PoolCluster) Replace(dst interface{}, opts ...Option) (Result, error) {
	s := c.NewSession()
	debugPrint("db: [session #%v] Replace", s.v)
	return s.Replace(dst, opts...)
}

//Delete delete record
func (c *PoolCluster) Delete(dst interface{}, opts ...Option) (Result, error) {
	s := c.NewSession()
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
func AddDb(driver, dsn string, opts ...DbOption) PoolClusterOpts {
	db := &dbEngine{
		Driver: driver,
		Dsn:    dsn,
		Opts:   opts,
	}
	return func(p *PoolCluster) *PoolCluster {
		p.pools = append(p.pools, db)
		return p
	}
}

//SetConnMaxLifetime ..
func SetConnMaxLifetime(d time.Duration) DbOption {
	return func(db *sql.DB) *sql.DB {
		db.SetConnMaxLifetime(d)
		return db
	}
}

//SetMaxIdleConns ..
func SetMaxIdleConns(n int) DbOption {
	return func(db *sql.DB) *sql.DB {
		db.SetMaxIdleConns(n)
		return db
	}
}

//SetMaxOpenConns ..
func SetMaxOpenConns(n int) DbOption {
	return func(db *sql.DB) *sql.DB {
		db.SetMaxOpenConns(n)
		return db
	}
}

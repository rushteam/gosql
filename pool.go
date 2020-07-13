package gosql

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"
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
		// debugPrint("db: Connect(%s,%s)", d.Driver, d.Dsn)
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
	vs           uint64
	pools        []*dbEngine
	forcePrimary bool
}

// PoolClusterOpts ..
type PoolClusterOpts func(p *PoolCluster) *PoolCluster

// Executor ..
func (c *PoolCluster) Executor(s *Session, primary bool) (*Session, error) {
	n := len(c.pools)
	if n == 0 {
		return nil, errors.New("not found db config")
	}
	if s == nil {
		s = &Session{v: atomic.AddUint64(&(c.vs), 1), ctx: context.Background()}
	}
	var dbx *dbEngine
	if primary || c.forcePrimary == true {
		//select primary db
		dbx = c.pools[0]
		debugPrint("db: [primary] dsn %s", dbx.Dsn)
	} else {
		//select replica db
		var i int
		if n > 1 {
			i = 1 + int(s.v)%(n-1)
		}
		dbx = c.pools[i]
		debugPrint("db: [replica#%d] %s", i, dbx.Dsn)
	}
	executor, err := dbx.Connect()
	if err != nil {
		return s, err
	}
	s.executor = executor
	return s, nil
}

//Primary select primary db
func (c *PoolCluster) Primary() (*Session, error) {
	return c.Executor(nil, true)
}

//Replica select replica db
func (c *PoolCluster) Replica(v uint64) (*Session, error) {
	return c.Executor(nil, false)
}

//Begin a transaction
func (c *PoolCluster) Begin() (*Session, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return s, err
	}
	executor, err := s.executor.(DB).Begin()
	if err != nil {
		return s, err
	}
	s.executor = executor
	return s, err
}

//QueryContext ..
func (c *PoolCluster) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	s, err := c.Executor(nil, false)
	if err != nil {
		return nil, err
	}
	return s.QueryContext(ctx, query, args...)
}

//Query ..
func (c *PoolCluster) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s, err := c.Executor(nil, false)
	if err != nil {
		return nil, err
	}
	return s.Query(query, args...)
}

//QueryRowContext ..
func (c *PoolCluster) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	s, _ := c.Executor(nil, false)
	return s.QueryRowContext(ctx, query, args...)
}

//QueryRow ..
func (c *PoolCluster) QueryRow(query string, args ...interface{}) *sql.Row {
	s, _ := c.Executor(nil, false)
	return s.QueryRow(query, args...)
}

//ExecContext ..
func (c *PoolCluster) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return s.ExecContext(ctx, query, args...)
}

//Exec ..
func (c *PoolCluster) Exec(query string, args ...interface{}) (sql.Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return s.Exec(query, args...)
}

//Fetch fetch record to model
func (c *PoolCluster) Fetch(dst interface{}, opts ...Option) error {
	s, err := c.Executor(nil, false)
	if err != nil {
		return err
	}
	return s.Fetch(dst, opts...)
}

//FetchAll fetch records to models
func (c *PoolCluster) FetchAll(dst interface{}, opts ...Option) error {
	s, err := c.Executor(nil, false)
	if err != nil {
		return err
	}
	return s.FetchAll(dst, opts...)
}

//Update update from model
func (c *PoolCluster) Update(dst interface{}, opts ...Option) (Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return s.Update(dst, opts...)
}

//Insert insert from model
func (c *PoolCluster) Insert(dst interface{}, opts ...Option) (Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return s.Insert(dst, opts...)
}

//Replace replace from model
func (c *PoolCluster) Replace(dst interface{}, opts ...Option) (Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
	return s.Replace(dst, opts...)
}

//Delete delete record
func (c *PoolCluster) Delete(dst interface{}, opts ...Option) (Result, error) {
	s, err := c.Executor(nil, true)
	if err != nil {
		return nil, err
	}
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

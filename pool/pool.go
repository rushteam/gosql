package pool

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"
)

//Cluster ..
type Cluster struct {
	dbType   string
	settings map[string]map[string][]string
	pool     map[string]*sql.DB
}

//Executor ..
type Executor interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
	//db
	// SetMaxIdleConns(n int)
	// SetMaxOpenConns(n int)
	// SetConnMaxLifetime(d time.Duration)
	// Stats() sql.DBStats
	// PingContext(ctx context.Context) error
	// Ping() error
	// Close() error
	// BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	// Begin() (*sql.Tx, error)
	// Driver() driver.Driver
	// Conn(ctx context.Context) (*sql.Conn, error)
	//tx
	// StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	// Stmt(stmt *sql.Stmt) *sql.Stmt
	// Commit() error
	// Rollback() error
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
	return c
}

//Start ..
func Start() {
	go func() {

	}()
}

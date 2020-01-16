package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

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
}

//Db ..
type Db interface {
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
	PingContext(ctx context.Context) error
	Ping() error
	Close() error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	Driver() driver.Driver
	Conn(ctx context.Context) (*sql.Conn, error)
}

//Tx ..
type Tx interface {
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	Stmt(stmt *sql.Stmt) *sql.Stmt
	Commit() error
	Rollback() error
}

//Cluster ..
type Cluster interface {
	// Open(name, node string) (*sql.DB, error)
	// Open(driverName string, dataSourceName string) (*DB, error)
	Master() (Executor, error)
	Slave() (Executor, error)
	Begin() (*sql.Tx, error)
	// Fetch(dst interface{}, opts ...builder.Option) error
	// FetchAll(dst interface{}, opts ...builder.Option) error
}

func debugPrint(format string, vals ...interface{}) {
	fmt.Printf(format+"\r\n", vals...)
}

package gosql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

//Debug env
var Debug = false

//ErrNoRows sql ErrNoRows
var ErrNoRows = sql.ErrNoRows

//Result sql Result
type Result sql.Result

//Executor is a Executor
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

//DB ..
type DB interface {
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
	Executor(s *Session, master bool) (Executor, error)
	Begin() (*Session, error)
	Fetch(dst interface{}, opts ...Option) error
	FetchAll(dst interface{}, opts ...Option) error
	Update(dst interface{}, opts ...Option) (Result, error)
	Insert(dst interface{}, opts ...Option) (Result, error)
	Replace(dst interface{}, opts ...Option) (Result, error)
	Delete(dst interface{}, opts ...Option) (Result, error)
}

func debugPrint(format string, vals ...interface{}) {
	if Debug {
		fmt.Printf(format+"\r\n", vals...)
	}
}

//Error ..
type Error struct {
	Number  uint16
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Number, e.Message)
}

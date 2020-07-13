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
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	Prepare(string) (*sql.Stmt, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	Exec(string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	QueryRow(string, ...interface{}) *sql.Row
}

//DB ..
type DB interface {
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	SetConnMaxLifetime(time.Duration)
	Stats() sql.DBStats
	PingContext(context.Context) error
	Ping() error
	Close() error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	Driver() driver.Driver
	Conn(context.Context) (*sql.Conn, error)
}

//Tx ..
type Tx interface {
	StmtContext(context.Context, *sql.Stmt) *sql.Stmt
	Stmt(*sql.Stmt) *sql.Stmt
	Commit() error
	Rollback() error
}

//Cluster ..
type Cluster interface {
	Executor(*Session, bool) (*Session, error)
	Begin() (*Session, error)
	Fetch(interface{}, ...Option) error
	FetchAll(interface{}, ...Option) error
	Update(interface{}, ...Option) (Result, error)
	Insert(interface{}, ...Option) (Result, error)
	Replace(interface{}, ...Option) (Result, error)
	Delete(interface{}, ...Option) (Result, error)
}

func debugPrint(format string, vals ...interface{}) {
	if Debug {
		fmt.Printf(format+"\r\n", vals...)
	}
}

// type Error mysql.MySQLError

//Error ..
type Error struct {
	Number  uint16
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Number, e.Message)
}

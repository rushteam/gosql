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
type Tx interface {
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	Stmt(stmt *sql.Stmt) *sql.Stmt
	Commit() error
	Rollback() error
}

//Session ..
type Session struct {
	ctx  context.Context
	exec Executor
	// clusterNode string
	// clusterName string
}

func newSession(ctx context.Context, exec Executor) *Session {
	return &Session{ctx, exec}
}

//Begin ..
func Begin() (*Session, error) {
	s := &Session{}
	return s, nil
}

//Commit Session
func (s *Session) Commit() error {
	if tx, ok := s.exec.(*sql.Tx); ok {
		return tx.Commit()
	}
	return fmt.Errorf("not found trans")
}

//Rollback Session
func (s *Session) Rollback() error {
	if tx, ok := s.exec.(*sql.Tx); ok {
		return tx.Rollback()
	}
	return fmt.Errorf("not found trans")
}

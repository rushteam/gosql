package db

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

var commonSession *Session

type executorFunc func(master bool) (Executor, error)

var vs uint64

//Session ..
type Session struct {
	master      bool
	ctx         context.Context
	getExecetor executorFunc
	v           uint64
}

//NewSession ..
func NewSession(ctx context.Context, master bool, getExecetor executorFunc) *Session {
	v := atomic.AddUint64(&vs, 1)
	return &Session{ctx: ctx, master: master, getExecetor: getExecetor, v: v}
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return err
	}
	// debugPrint("db: [sql] %s %v", sql, args)
	debugPrint("db: [session #%v] Fetch", s.v)
	rows, err := executor.QueryContext(s.ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.Scan(rows, dst)
}

//FetchAll ..
func (s *Session) FetchAll(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	sql, args := builder.Select(
		builder.Table(dstStruct.TableName()),
	)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return err
	}
	debugPrint("db: [session #%v] FetchAll", s.v)
	rows, err := executor.QueryContext(s.ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, dst)
}

//Commit ..
func (s *Session) Commit() error {
	debugPrint("db: [session #%v] Commit", s.v)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return err
	}
	return executor.(*sql.Tx).Commit()
}

//Begin ..
func Begin() (*Session, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	debugPrint("db: [session #%v] Begin", commonSession.v)
	executor, err := commonSession.getExecetor(true)
	if err != nil {
		return nil, err
	}
	getExecetor := func(master bool) (Executor, error) {
		return executor.(DB).Begin()
	}
	return NewSession(context.TODO(), true, getExecetor), nil
}

//Fetch ..
func Fetch(dst interface{}, opts ...builder.Option) error {
	if commonSession == nil {
		return errors.New("db: not found session")
	}
	return commonSession.Fetch(dst, opts...)
}

//FetchAll ..
func FetchAll(dst interface{}, opts ...builder.Option) error {
	if commonSession == nil {
		return errors.New("db: not found session")
	}
	return commonSession.FetchAll(dst, opts...)
}

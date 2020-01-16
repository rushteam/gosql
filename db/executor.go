package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

var commonSession *Session

type executorFunc func(master bool) (Executor, error)

//Session ..
type Session struct {
	master      bool
	ctx         context.Context
	getExecetor executorFunc
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
	debugPrint("db: [sql] %s %v", sql, args)
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
	rows, err := executor.QueryContext(s.ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, dst)
}

//Commit ..
func (s *Session) Commit() error {
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
	getExecetor := func(master bool) (Executor, error) {
		executor, err := commonSession.getExecetor(true)
		if err != nil {
			return nil, err
		}
		return executor.(DB).Begin()
	}
	return &Session{master: true, ctx: context.TODO(), getExecetor: getExecetor}, nil
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

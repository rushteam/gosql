package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

var commonSession *Session

//Session ..
type Session struct {
	master  bool
	cluster Cluster
	ctx     context.Context
}

//getExcetor ..
func (s *Session) getExcetor() (Executor, error) {
	if s.master == true {
		return s.cluster.Master()
	}
	return s.cluster.Slave()
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	executor, err := s.getExcetor()
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
	executor, err := s.getExcetor()
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
	excetor, err := s.getExcetor()

	excetor.(*sql.DB).()
}

//Begin ..
func Begin() *Session {
	return &Session{cluster: commonSession.cluster, master: true, ctx: context.TODO()}
}

//Fetch ..
func Fetch(dst interface{}, opts ...builder.Option) error {
	if commonSession == nil {
		return errors.New("not found session")
	}
	return commonSession.Fetch(dst, opts...)
}

//FetchAll ..
func FetchAll(dst interface{}, opts ...builder.Option) error {
	if commonSession == nil {
		return errors.New("not found session")
	}
	return commonSession.FetchAll(dst, opts...)
}

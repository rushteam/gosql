package db

import (
	"context"
	"errors"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

var commonSession *Session

type Session struct {
	readOnly bool
	cluster  Cluster
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	ctx := context.TODO()
	executor, err := s.cluster.Slave()
	if err != nil {
		return err
	}
	rows, err := executor.QueryContext(ctx, sql, args...)
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
	ctx := context.TODO()
	executor, err := s.cluster.Slave()
	if err != nil {
		return err
	}
	rows, err := executor.QueryContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, dst)
}

// func Begin() *Session {
// 	commonSession = &Session{cluster: cluster, readOnly: false}
// 	return commonSession
// }

func Master() *Session {
	return &Session{cluster: cluster, readOnly: false}
}

// func Slave() *Session {
// 	commonSession = &Session{cluster: cluster, readOnly: true}
// 	return commonSession
// }
func Init(cluster Cluster) {
	commonSession = &Session{cluster: cluster}
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

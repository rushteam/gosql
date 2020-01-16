package db

import (
	"context"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

//Fetch ..
func Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	engine, err := cluster.Slave()
	if err != nil {
		return err
	}
	ctx := context.TODO()
	rows, err := engine.QueryContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.Scan(rows, dst)
}

//FetchAll ..
func FetchAll(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	sql, args := builder.Select(
		builder.Table(dstStruct.TableName()),
	)
	engine, err := cluster.Slave()
	if err != nil {
		return err
	}
	ctx := context.TODO()
	rows, err := engine.QueryContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, dst)
}

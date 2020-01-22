package db

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

//Result ..
type Result sql.Result

//自动更新时间
var autoFillCreatedAtAndUpdatedAtField = true
var createdAtField = "created_at"
var updatedAtField = "updated_at"
var deletedAtField = "deleted_at"

var commonSession *Session

type executorFunc func(master bool) (Executor, error)

var vs uint64

//Session ..
type Session struct {
	master      bool
	ctx         context.Context
	getExecetor executorFunc
	done        int32
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
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
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

//Update ..
func (s *Session) Update(dst interface{}, opts ...builder.Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveModelToMap(dst)
	if err != nil {
		return nil, err
		// panic(err)
	}
	pk := dstStruct.GetPk()
	if pk != "" {
		//若主键值不为空则增加主键条件
		if id, ok := fields[pk]; ok {
			if id != "" && id != nil {
				opts = append(opts, builder.Where(pk, id))
			}
			// delete(fields, pk)
		}
	}
	updateFields := make(map[string]interface{}, 0)
	for k, v := range fields {
		if k == pk || k == "" {
			continue
		}
		//过滤掉 v 是空的值 todo 指针怎么办?
		if v == nil || v == "" {
			continue
		}
		//过滤掉 model 中的主键 防止修改
		// if pk != "" && k == pk {
		// 	continue
		// }
		updateFields[k] = v
	}
	//若开启自动填充时间，则尝试自动填充时间
	if autoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间
		updateFields[updatedAtField] = time.Now()
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	//todo 这里增加批量操作直接setMap(updateFields)
	for k, v := range updateFields {
		opts = append(opts, builder.Set(k, v))
	}

	sql, args := builder.Update(opts...)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return nil, err
	}
	rst, err := executor.ExecContext(s.ctx, sql, args...)
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//将数据更新到结构体上
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Commit ..
func (s *Session) Commit() error {
	if atomic.LoadInt32(&s.done) == 1 {
		return errors.New("db: [] has done")
	}
	debugPrint("db: [session #%v] Commit", s.v)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return err
	}
	return executor.(*sql.Tx).Commit()
}

//Rollback ..
func (s *Session) Rollback() error {
	debugPrint("db: [session #%v] Rollback", s.v)
	executor, err := s.getExecetor(s.master)
	if err != nil {
		return err
	}
	return executor.(*sql.Tx).Rollback()
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

//Update ..
func Update(dst interface{}, opts ...builder.Option) (Result, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	return commonSession.Update(dst, opts...)
}

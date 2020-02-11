package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"
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
	ctx      context.Context
	done     int32
	v        uint64
	executor Executor
	mutex    sync.RWMutex
	// getExecetor executorFunc
	// master      bool
	cluster Cluster
}

//NewSession ..
func NewSession(ctx context.Context, c Cluster) *Session {
	v := atomic.AddUint64(&vs, 1)
	return &Session{ctx: ctx, cluster: c, v: v}
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	executor, err := s.cluster.Master()
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
	executor, err := s.cluster.Master()
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
	executor, err := s.cluster.Master()
	if err != nil {
		return nil, err
	}
	rst, err := executor.ExecContext(s.ctx, sql, args...)
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//将数据更新到结构体上
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Insert ..
func (s *Session) Insert(dst interface{}, opts ...builder.Option) (Result, error) {
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
	updateFields := make(map[string]interface{}, 0)
	for k, v := range fields {
		if k == pk || k == "" {
			continue
		}
		updateFields[k] = v
	}
	//若开启自动填充时间，则尝试自动填充时间
	if autoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间/创建时间
		updateFields[updatedAtField] = time.Now()
		updateFields[createdAtField] = time.Now()
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	//todo 这里增加批量操作直接setMap(updateFields)
	for k, v := range updateFields {
		opts = append(opts, builder.Set(k, v))
	}

	sql, args := builder.Insert(opts...)
	executor, err := s.cluster.Master()
	if err != nil {
		return nil, err
	}
	rst, err := executor.ExecContext(s.ctx, sql, args...)
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//将数据更新到结构体上
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Replace ..
func (s *Session) Replace(dst interface{}, opts ...builder.Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveModelToMap(dst)
	if err != nil {
		return nil, err
		// panic(err)
	}
	updateFields := make(map[string]interface{}, 0)
	for k, v := range fields {
		updateFields[k] = v
	}
	//若开启自动填充时间，则尝试自动填充时间
	if autoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间/创建时间
		updateFields[updatedAtField] = time.Now()
		updateFields[createdAtField] = time.Now()
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	//todo 这里增加批量操作直接setMap(updateFields)
	for k, v := range updateFields {
		opts = append(opts, builder.Set(k, v))
	}

	sql, args := builder.Replace(opts...)
	executor, err := s.cluster.Master()
	if err != nil {
		return nil, err
	}
	rst, err := executor.ExecContext(s.ctx, sql, args...)
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//将数据更新到结构体上
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Delete ..
func (s *Session) Delete(dst interface{}, opts ...builder.Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveModelToMap(dst)
	if err != nil {
		return nil, err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	pk := dstStruct.GetPk()
	for k, v := range fields {
		//仅仅取model中的pk，其他一律忽略
		if k == pk && k != "" {
			opts = append(opts, builder.Where(k, v))
		}
	}
	sql, args := builder.Delete(opts...)
	executor, err := s.cluster.Master()
	if err != nil {
		return nil, err
	}
	rst, err := executor.ExecContext(s.ctx, sql, args...)
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	return rst, err
}

//Commit ..
func (s *Session) Commit() error {
	debugPrint("db: [session #%v] Commit", s.v)
	if s.executor == nil {
		return errors.New("not found trans")
	}
	return s.executor.(*sql.Tx).Commit()
}

//Rollback ..
func (s *Session) Rollback() error {
	debugPrint("db: [session #%v] Rollback", s.v)
	if s.executor == nil {
		return errors.New("not found trans")
	}
	return s.executor.(*sql.Tx).Commit()
}

//Begin ..
func Begin() (*Session, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	s := NewSession(context.TODO(), commonSession.cluster)
	debugPrint("db: [session #%v] Begin", commonSession.v)
	executor, err := commonSession.cluster.Master()
	if err != nil {
		return nil, err
	}
	executor, err = executor.(DB).Begin()
	if err != nil {
		return nil, err
	}
	s.executor = executor
	return s, nil
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

//Insert ..
func Insert(dst interface{}, opts ...builder.Option) (Result, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	return commonSession.Insert(dst, opts...)
}

//Replace ..
func Replace(dst interface{}, opts ...builder.Option) (Result, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	return commonSession.Replace(dst, opts...)
}

//Delete ..
func Delete(dst interface{}, opts ...builder.Option) (Result, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	return commonSession.Delete(dst, opts...)
}

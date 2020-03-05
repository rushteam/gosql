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
	// master      bool
	cluster Cluster
}

//NewSession ..
func NewSession(ctx context.Context, c Cluster) *Session {
	v := atomic.AddUint64(&vs, 1)
	return &Session{ctx: ctx, cluster: c, v: v}
}

//Executor ..
func (s *Session) Executor(master bool) (Executor, error) {
	var err error
	if s.executor == nil {
		if master == true {
			s.executor, err = s.cluster.Master()
		} else {
			s.executor, err = s.cluster.Slave()
		}
	}
	return s.executor, err
}

//QueryContext ..
func (s *Session) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//todo 增加强制master逻辑
	db, err := s.Executor(false)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args)
}

//Query ..
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.QueryContext(s.ctx, query, args)
}

//QueryRowContext ..
func (s *Session) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	//todo 增加强制master逻辑
	db, err := s.Executor(false)
	if err != nil {
		return nil
	}
	return db.QueryRowContext(ctx, query, args)
}

//QueryRow ..
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.QueryRowContext(s.ctx, query, args)
}

//ExecContext ..
func (s *Session) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	debugPrint("db: [session #%v] %s %v", s.v, sql, args)
	db, err := s.Executor(true)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args)
}

//Exec ..
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.ExecContext(s.ctx, query, args)
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...builder.Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, builder.Table(dstStruct.TableName()))
	sql, args := builder.Select(opts...)
	rows, err := s.QueryContext(s.ctx, sql, args...)
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
	rows, err := s.QueryContext(s.ctx, sql, args...)
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
	rst, err := s.ExecContext(s.ctx, sql, args...)
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
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//将数据更新到结构体上
	if err == nil {
		updateFields[pk], _ = rst.LastInsertId()
	}
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Replace ..
func (s *Session) Replace(dst interface{}, opts ...builder.Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	pk := dstStruct.GetPk()
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
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//将数据更新到结构体上
	if err == nil {
		updateFields[pk], _ = rst.LastInsertId()
	}
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
		if k == pk && k != "" {
			//仅仅取model中的pk，其他一律忽略
			opts = append(opts, builder.Where(k, v))
			break
		} else {
			opts = append(opts, builder.Where(k, v))
		}
	}
	sql, args := builder.Delete(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
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
	return s.executor.(*sql.Tx).Rollback()
}

//Begin ..
func Begin() (*Session, error) {
	if commonSession == nil {
		return nil, errors.New("db: not found session")
	}
	s := NewSession(context.TODO(), commonSession.cluster)
	debugPrint("db: [session #%v] Begin", s.v)
	executor, err := s.Executor(true)
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

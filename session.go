package gosql

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/rushteam/gosql/scanner"
)

//自动更新时间
var autoFillCreatedAtAndUpdatedAtField = true
var createdAtField = "created_at"
var updatedAtField = "updated_at"
var deletedAtField = "deleted_at"

//SessionOpts ..
type SessionOpts func(s *Session) *Session

//Session ..
type Session struct {
	cluster     Cluster
	ctx         context.Context
	done        int32
	v           uint64
	executor    Executor
	mutex       sync.RWMutex
	forceMaster bool
}

//Master 强制master
func (s *Session) Master() *Session {
	s.forceMaster = true
	return s
}

//Executor ..
func (s *Session) Executor(master bool) (Executor, error) {
	var err error
	if s.executor == nil {
		s.executor, err = s.cluster.Executor(s, master)
	}
	return s.executor, err
}

//QueryContext ..
func (s *Session) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := s.Executor(false)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

//Query ..
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.QueryContext(s.ctx, query, args...)
}

//QueryRowContext ..
func (s *Session) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := s.Executor(false)
	if err != nil {
		row := &sql.Row{}
		rowErr := (*error)(unsafe.Pointer(row))
		*rowErr = err
		return row
	}
	return db.QueryRowContext(ctx, query, args...)
}

//QueryRow ..
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.QueryRowContext(s.ctx, query, args...)
}

//ExecContext ..
func (s *Session) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	debugPrint("db: [session #%v] %s %v", s.v, query, args)
	db, err := s.Executor(true)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

//Exec ..
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.ExecContext(s.ctx, query, args...)
}

//Fetch ..
func (s *Session) Fetch(dst interface{}, opts ...Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, Table(dstStruct.TableName()))
	sql, args := SelectSQL(opts...)
	rows, err := s.QueryContext(s.ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.Scan(rows, dst)
}

//FetchAll ..
func (s *Session) FetchAll(dst interface{}, opts ...Option) error {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	opts = append(opts, Table(dstStruct.TableName()))
	sql, args := SelectSQL(opts...)
	rows, err := s.QueryContext(s.ctx, sql, args...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, dst)
}

//Update ..
func (s *Session) Update(dst interface{}, opts ...Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveStructValue(dst)
	if err != nil {
		return nil, err
	}
	pk := dstStruct.GetPk()
	if pk != "" {
		//若主键值不为空则增加主键条件
		if id, ok := fields[pk]; ok {
			if id != "" && id != nil {
				opts = append(opts, Where(pk, id))
			}
			// delete(fields, pk)
		}
	}
	updateFields := make(map[string]interface{}, 0)
	for k, v := range fields {
		if k == pk || k == "" {
			continue
		}
		//过滤掉 v 是空的值 todo 会出现指针吗?要是指针怎么处理?
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
	opts = append(opts, Table(dstStruct.TableName()))
	opts = append(opts, Params(updateFields))
	// for k, v := range updateFields {
	// 	opts = append(opts, Set(k, v))
	// }
	sql, args := UpdateSQL(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//将数据更新到结构体上
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Insert ..
func (s *Session) Insert(dst interface{}, opts ...Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveStructValue(dst)
	if err != nil {
		return nil, err
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
	opts = append(opts, Table(dstStruct.TableName()))
	opts = append(opts, Params(updateFields))
	// for k, v := range updateFields {
	// 	opts = append(opts, Set(k, v))
	// }
	sql, args := InsertSQL(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//将数据更新到结构体上
	if err == nil {
		updateFields[pk], _ = rst.LastInsertId()
	}
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Replace ..
func (s *Session) Replace(dst interface{}, opts ...Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	pk := dstStruct.GetPk()
	fields, err := scanner.ResolveStructValue(dst)
	if err != nil {
		return nil, err
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
	opts = append(opts, Table(dstStruct.TableName()))
	opts = append(opts, Params(updateFields))
	// for k, v := range updateFields {
	// 	opts = append(opts, Set(k, v))
	// }
	sql, args := ReplaceSQL(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//将数据更新到结构体上
	if err == nil {
		updateFields[pk], _ = rst.LastInsertId()
	}
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Delete ..
func (s *Session) Delete(dst interface{}, opts ...Option) (Result, error) {
	dstStruct, err := scanner.ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	fields, err := scanner.ResolveStructValue(dst)
	if err != nil {
		return nil, err
	}
	opts = append(opts, Table(dstStruct.TableName()))
	pk := dstStruct.GetPk()
	for k, v := range fields {
		if k == pk && k != "" {
			//仅仅取model中的pk，其他一律忽略
			opts = append(opts, Where(k, v))
			break
		} else {
			opts = append(opts, Where(k, v))
		}
	}
	sql, args := DeleteSQL(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
	return rst, err
}

//begin
func (s *Session) begin() error {
	debugPrint("db: [session #%v] Begin", s.v)
	executor, err := s.cluster.Executor(s, true)
	if err != nil {
		return err
	}
	executor, err = executor.(DB).Begin()
	if err != nil {
		return err
	}
	s.executor = executor
	return nil
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

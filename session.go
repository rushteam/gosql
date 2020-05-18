package gosql

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/rushteam/gosql/scanner"
)

//AutoFillCreatedAtAndUpdatedAtField 自动更新时间
var AutoFillCreatedAtAndUpdatedAtField = false

//AutoFieldCreatedAt when insert auto set time
var AutoFieldCreatedAt = "created_at"

//AutoFieldUpdatedAt when update auto set time
var AutoFieldUpdatedAt = "updated_at"

//todo soft delete
// var AutoFieldDeletedAt = "deleted_at"

//Session ..
type Session struct {
	v        uint64
	executor Executor
	mutex    sync.RWMutex
	done     int32
	ctx      context.Context
}

//Executor ..
func (s *Session) Executor() (Executor, error) {
	var err error
	if s.executor == nil {
		return nil, errors.New("not found db")
	}
	return s.executor, err
}

//QueryContext ..
func (s *Session) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	debugPrint("db: [session #%v] Query %s %v", s.v, query, args)
	db, err := s.Executor()
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
	debugPrint("db: [session #%v] Query %s %v", s.v, query, args)
	db, _ := s.Executor()
	// db, err := s.Executor()
	// if err != nil {
	// 	row := &sql.Row{}
	// 	rowErr := (*error)(unsafe.Pointer(row))
	// 	*rowErr = err
	// 	return row
	// }
	return db.QueryRowContext(ctx, query, args...)
}

//QueryRow ..
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.QueryRowContext(s.ctx, query, args...)
}

//ExecContext ..
func (s *Session) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	debugPrint("db: [session #%v] Exec %s %v", s.v, query, args)
	db, err := s.Executor()
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
	debugPrint("db: [session #%v] Fetch()", s.v)
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
	debugPrint("db: [session #%v] FetchAll()", s.v)
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
	debugPrint("db: [session #%v] Update", s.v)
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
		}
	}
	updateFields := make(map[string]interface{}, 0)
	for k, v := range fields {
		if k == "" || k == pk {
			continue
		}
		//过滤掉 v 是空的值 todo 会出现指针吗?要是指针怎么处理?
		if v == nil || v == "" {
			continue
		}
		updateFields[k] = v
	}
	//若开启自动填充时间，则尝试自动填充时间
	if AutoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间
		updateFields[AutoFieldUpdatedAt] = time.Now()
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
	debugPrint("db: [session #%v] Insert", s.v)
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
		//skip pk
		if k == "" || k == pk {
			continue
		}
		updateFields[k] = v
	}
	//若开启自动填充时间，则尝试自动填充时间
	if AutoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间/创建时间
		updateFields[AutoFieldUpdatedAt] = time.Now()
		updateFields[AutoFieldCreatedAt] = time.Now()
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
		if pk != "" {
			updateFields[pk], _ = rst.LastInsertId()
		}
	}
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Replace ..
func (s *Session) Replace(dst interface{}, opts ...Option) (Result, error) {
	debugPrint("db: [session #%v] Replace", s.v)
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
	if AutoFillCreatedAtAndUpdatedAtField == true {
		//强制填充更新时间/创建时间
		updateFields[AutoFieldUpdatedAt] = time.Now()
		updateFields[AutoFieldCreatedAt] = time.Now()
	}
	opts = append(opts, Table(dstStruct.TableName()))
	opts = append(opts, Params(updateFields))
	// for k, v := range updateFields {
	// 	opts = append(opts, Set(k, v))
	// }
	sql, args := ReplaceSQL(opts...)
	rst, err := s.ExecContext(s.ctx, sql, args...)
	//update model pk
	if err == nil {
		if pk != "" {
			updateFields[pk], _ = rst.LastInsertId()
		}
	}
	scanner.UpdateModel(dst, updateFields)
	return rst, err
}

//Delete ..
func (s *Session) Delete(dst interface{}, opts ...Option) (Result, error) {
	debugPrint("db: [session #%v] Delete", s.v)
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
		if k != "" && k == pk {
			//just use pk,igone other case
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

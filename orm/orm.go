package orm

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/db"
	"github.com/mlboy/godb/scanner"
)

//BuilderHandler ..
type BuilderHandler func(*builder.SQLSegments)

var autoFillCreatedAndUpdatedField = true
var createdAtField = "created_at"
var updatedAtField = "updated_at"
var deletedAtField = "deleted_at"

var cluster db.Cluster

//Init ..
func Init(c db.Cluster) {
	cluster = c
}

//Model 加载模型 orm.Model(&tt{}).Builder(func(){}).Find()
func Model(dst interface{}) *ORM {
	o := &ORM{}
	o.Model(dst)
	return o
}

//ORM ..
type ORM struct {
	dst         interface{}
	builder     *builder.SQLSegments
	modelStruct *scanner.StructData
	fields      map[string]interface{}
	ctx         context.Context
	clusterNode string
	executor    db.Executor
}

//Model 初始化
func (o *ORM) Model(dst interface{}) *ORM {
	var err error
	o.dst = dst
	//解析结构体
	o.modelStruct, err = scanner.ResolveModelStruct(o.dst)
	if err != nil {
		panic(err)
	}
	o.fields, err = scanner.ResolveModelToMap(o.dst)
	if err != nil {
		panic(err)
	}
	o.builder = builder.New()
	//获取表名
	o.builder.Table(o.modelStruct.TableName())
	o.clusterNode = "slave" //salver
	o.ctx = context.Background()
	return o
}

//Db ..
func (o *ORM) Db() (db.Executor, error) {
	if o.executor != nil {
		return o.executor, nil
	}
	return cluster.Slave()

}

//Begin ..
func Begin() (*ORM, error) {
	o := &ORM{}
	var err error
	o.executor, err = cluster.Master()
	if err != nil {
		return nil, err
	}
	if db, ok := o.executor.(*sql.DB); ok {
		o.executor, err = db.Begin()
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

//Commit ..
func (o *ORM) Commit() error {
	if tx, ok := o.executor.(*sql.Tx); ok {
		return tx.Commit()
	}
	return errors.New("db: commit fail, not found trans")
}

//Rollback ..
func (o *ORM) Rollback() error {
	if tx, ok := o.executor.(*sql.Tx); ok {
		return tx.Rollback()
	}
	return errors.New("db: rollback fail, not found trans")
}

//Query ..
func (o *ORM) Query(query string, args ...interface{}) (*sql.Rows, error) {
	excutor, err := o.Db()
	// excutor, err := o.cluster.Db(o.clusterName, o.clusterNode)
	if err != nil {
		return nil, err
	}
	return excutor.QueryContext(o.ctx, query, args...)
}

//Exec ..
func (o *ORM) Exec(query string, args ...interface{}) (sql.Result, error) {
	excutor, err := o.Db()
	// excutor, err := o.cluster.Db(o.clusterName, o.clusterNode)
	if err != nil {
		return nil, err
	}
	return excutor.ExecContext(o.ctx, query, args...)
}

//Master 强制master
func (o *ORM) Master() *ORM {
	o.executor, _ = cluster.Master()
	return o
}

//Fetch 拉取
func (o *ORM) Fetch() error {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Find() ")
	}
	o.builder.Limit(1)
	rows, err := o.Query(o.builder.BuildSelect(), o.builder.Args()...)
	if err != nil {
		return err
	}
	return scanner.Scan(rows, o.dst)
}

//FetchAll 查找数据
func (o *ORM) FetchAll() error {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Find() ")
	}
	rows, err := o.Query(o.builder.BuildSelect(), o.builder.Args()...)
	if err != nil {
		return err
	}
	return scanner.ScanAll(rows, o.dst)
}

/*
Limit 限制数
*/
func (o *ORM) Limit(n int) *ORM {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Limit() ")
	}
	o.builder.Limit(n)
	return o
}

/*
Offset 偏移量
*/
func (o *ORM) Offset(n int) *ORM {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Offset() ")
	}
	o.builder.Offset(n)
	return o
}

/*
Where 条件
*/
func (o *ORM) Where(key interface{}, vals ...interface{}) *ORM {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Where() ")
	}
	// if len(vals) > 0 {
	// 	o.fields[key] = vals[0]
	// }
	o.builder.Where(key, vals...)
	return o
}

/*
UpdateField 更新字段
*/
func (o *ORM) UpdateField(k string, v interface{}) *ORM {
	r, _ := regexp.Compile(`\[(\+|\-)\]?([a-zA-Z0-9_.\-\=\s\?\(\)]*)`)
	match := r.FindStringSubmatch(k)
	key := scanner.FormatName(match[2])
	if match[1] == "" {
		o.fields[key] = v
	} else {
		//todo [opt]1 这种数据时候的处理
	}
	o.builder.UpdateField(k, v)
	return o
}

/*
Update 更新数据
*/
func (o *ORM) Update(fs ...BuilderHandler) (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Update() ")
	}
	// list, err := scanner.ResolveModelToMap(o.dst)
	if len(fs) > 0 {
		for _, f := range fs {
			f(o.builder)
		}
	}
	// pk := o.modelStruct.GetPk()
	// if pk != "" {
	// 	if id, ok := list[pk]; ok {
	// 		o.Where(pk, id)
	// 		delete(list, pk)
	// 	}
	// }
	if autoFillCreatedAndUpdatedField == true {
		if _, ok := o.fields[updatedAtField]; !ok {
			o.fields[updatedAtField] = time.Now()
		}
	}
	o.builder.Update(o.fields)
	rst, err := o.Exec(o.builder.BuildUpdate(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	scanner.UpdateModel(o.dst, o.fields)
	return rst, nil
}

//Insert 插入数据
func (o *ORM) Insert() (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Insert() ")
	}
	// list, err := scanner.ResolveModelToMap(o.dst)
	if autoFillCreatedAndUpdatedField == true {
		if _, ok := o.fields[createdAtField]; !ok {
			o.fields[createdAtField] = time.Now()
		}
		if _, ok := o.fields[updatedAtField]; !ok {
			o.fields[updatedAtField] = time.Now()
		}
	}
	o.builder.Insert(o.fields)
	rst, err := o.Exec(o.builder.BuildInsert(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	pk := o.modelStruct.GetPk()
	if id, err := rst.LastInsertId(); err == nil && pk != "" {
		o.fields[pk] = id
	}
	scanner.UpdateModel(o.dst, o.fields)
	return rst, nil
}

//Delete 插入数据
func (o *ORM) Delete() (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Delete() ")
	}
	o.builder.Delete()
	rst, err := o.Exec(o.builder.BuildDelete(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	return rst, nil
}

//BuilderUpdate ..
// func (o *ORM) BuilderUpdate(f BuilderHandler) (sql.Result, error) {
// 	f(o.builder)
// 	rst, err := o.Db().Exec(o.builder.BuildUpdate(), o.builder.Args()...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rst, nil
// }

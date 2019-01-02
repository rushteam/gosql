package orm

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/mlboy/godb/builder"
	"github.com/mlboy/godb/scanner"
)

var defaultDb *sql.DB

//InitDefaultDb 设置默认db
func InitDefaultDb(db *sql.DB) {
	defaultDb = db
}

//ORM ..
type ORM struct {
	db          *sql.DB
	dst         interface{}
	builder     *builder.SQLSegments
	modelStruct *scanner.StructData
}

//Model 加载模型 orm.Model(&tt{}).Builder(func(){}).Find()
func Model(dst interface{}) *ORM {
	var err error
	o := &ORM{}
	o.db = defaultDb
	o.dst = dst
	o.modelStruct, err = scanner.ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		panic(err)
	}
	o.builder = builder.New()
	name, err := scanner.ResolveModelTableName(dst)
	if err != nil {
		panic(err)
	}
	o.builder.Table(name)
	return o
}

//Db ..
func (o *ORM) Db() *sql.DB {
	if o.db == nil {
		panic("orm: not found db, must init a db first")
	}
	if o.builder == nil {
		panic("orm: must call Model() first, before call Db() ")
	}
	return o.db
}

//Fetch 拉取
func (o *ORM) Fetch() error {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Find() ")
	}
	list, err := scanner.ResolveModelToMap(o.dst)
	if err != nil {
		return err
	}
	for k, v := range list {
		fmt.Println(k, v)
	}

	return o.Find()
}

/*
Find 查找数据
*/
func (o *ORM) Find() error {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Find() ")
	}
	o.builder.Limit(1)
	rows, err := o.Db().Query(o.builder.BuildSelect(), o.builder.Args()...)
	if err != nil {
		return err
	}
	return scanner.Scan(rows, o.dst)
}

/*
FindAll 查找数据
*/
func (o *ORM) FindAll() error {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Find() ")
	}
	rows, err := o.Db().Query(o.builder.BuildSelect(), o.builder.Args()...)
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
	o.builder.Where(key, vals...)
	return o
}

/*
Update 更新数据
*/
func (o *ORM) Update(fs ...BuilderHandler) (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Update() ")
	}
	list, err := scanner.ResolveModelToMap(o.dst)
	if err != nil {
		return nil, err
	}
	pk := o.modelStruct.GetPk()
	if pk != "" {
		if id, ok := list[pk]; ok {
			o.Where(pk, id)
			delete(list, pk)
		}
	}
	o.builder.Update(list)
	rst, err := o.Db().Exec(o.builder.BuildUpdate(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	// o.modelStruct.GetStructField("").Index()
	// if id, err := rst.LastInsertId(); err != nil {
	// 	// list[pk] = id
	// 	scanner.UpdateModel(o.dst, list)
	// }
	scanner.UpdateModel(o.dst, list)
	return rst, nil
}

//Insert 插入数据
func (o *ORM) Insert() (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Update() ")
	}
	pk := o.modelStruct.GetPk()
	list, err := scanner.ResolveModelToMap(o.dst)
	if err != nil {
		return nil, err
	}
	o.builder.Insert(list)
	rst, err := o.Db().Exec(o.builder.BuildInsert(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	if id, err := rst.LastInsertId(); err == nil && pk != "" {
		list[pk] = id
	}
	scanner.UpdateModel(o.dst, list)
	return rst, nil
}

//Delete 插入数据
func (o *ORM) Delete() (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Update() ")
	}
	o.builder.Delete()
	rst, err := o.Db().Exec(o.builder.BuildDelete(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	return rst, nil
}

//BuilderHandler ..
type BuilderHandler func(*builder.SQLSegments)

//BuilderUpdate ..
func (o *ORM) BuilderUpdate(f BuilderHandler) (sql.Result, error) {
	f(o.builder)
	rst, err := o.Db().Exec(o.builder.BuildUpdate(), o.builder.Args()...)
	if err != nil {
		return nil, err
	}
	return rst, nil
}

package orm

import (
	"database/sql"
	"reflect"

	"../builder"
	"../scanner"
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
func (o *ORM) Update() (sql.Result, error) {
	if o.builder == nil {
		panic("orm: must call Model() first, before call Update() ")
	}
	pk := o.modelStruct.GetPk()
	list, err := scanner.ResolveModelToMap(o.dst)
	if err != nil {
		return nil, err
	}
	if id, ok := list[pk]; ok {
		o.Where(pk, id)
		delete(list, pk)
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
	if id, err := rst.LastInsertId(); err != nil {
		list[pk] = id
	}
	// o.modelStruct.GetStructField("").Index()
	// if id, err := rst.LastInsertId(); err != nil {
	// 	// list[pk] = id
	// 	scanner.UpdateModel(o.dst, list)
	// }
	scanner.UpdateModel(&o.dst, list)
	return rst, nil
}

// s := builder.New()
// s.Table("tbl1.t1")
// s.Flag("DISTANCE")
// s.Field("*")
// s.Table("tbl1.t1")
// s.Where("t1.status", "0")
// s.Where("type", "A")
// s.Where("[in]sts", []string{"1", "2", "3", "4"})
// s.Where("[in]sts2", 1)
// s.Where(func(s *builder.Clause) {
// 	s.Where("a", "200")
// 	s.Where("b", "100")
// })
// s.Where("aaa = 999")
// s.Where("[#]ccc = ?", 888)
// s.Join("tbl3", "a", "=", "b")
// s.Having("ss", "1")
// s.Where("[~]a", "AA")
// s.Where("[exists]", "AA")
// s.Where("[exists]", func(s *builder.SQLSegments) {
// 	s.Where("xx", 10000)
// })
// s.GroupBy("id")
// s.OrderBy("id desc", "id asc")
// s.Limit(30)
// s.Offset(10)
// s.ForUpdate()

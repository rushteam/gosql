package orm

import (
	"database/sql"
	"reflect"

	"../builder"
	"../scanner"
)

//ORM ..
type ORM struct {
	db         *sql.DB
	builder    *builder.SQLSegments
	structData *scanner.StructData
}

//Model 加载模型 orm.Model(&tt{}).Builder(func(){}).Find()
func Model(dst interface{}) *ORM {
	var err error
	o := &ORM{}
	o.builder = builder.New()
	o.structData, err = scanner.ResolveStruct(reflect.TypeOf(dst))
	if err != nil {
		panic(err)
	}
	o.builder.Table(o.structData.TableName)
	return o
}

/*
Find 查找数据
*/
func (o *ORM) Find() *ORM {
	if o.builder == nil {
		panic("must call Model() first, before call Find() ")
	}
	o.builder.Limit(1)
	o.builder.BuildSelect()
	return o
}

/*
Where 条件
*/
func (o *ORM) Where(key interface{}, vals ...interface{}) *ORM {
	if o.builder == nil {
		panic("must call Model() first, before call Where() ")
	}
	o.builder.Where(key, vals...)
	return o
}

/*
Update 更新数据
*/
func (o *ORM) Update() *ORM {
	if o.builder == nil {
		panic("must call Model() first, before call Update() ")
	}
	o.builder.BuildUpdate()
	return o
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
func (o *ORM) FindAll() {

}

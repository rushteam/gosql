package main

import (
	"fmt"
	"time"

	"github.com/mlboy/godb/builder"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mlboy/godb/db"
	"github.com/mlboy/godb/orm"
)

type S struct {
	Key string `db:"key"`
	Val string `db:"val"`
}
type T struct {
	ID      int64   `db:"id,pk"`
	Uid     string  `db:",index"`
	Typ     *string `db:"typ,index"`
	Expires int     `db:"expires"`
	XX      int     `db:"-"`
	Scope   string  `db:"scope,csv"`
	// Scope     json.RawMessage `db:"scope,csv"`
	UpdatedAt *time.Time `db:"updated_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func (t T) TableName() string {
	return "login"
}
func main() {
	//id type client_id client_secret salt created updated metadata

	// orm.Db.(orm.Select(&T).Where())
	// err := orm.Model().Where("id", 1).Find(&t)
	// err := orm.Model().Where("id", 1).Update(&t)
	// err := orm.Model().Insert(&t)
	// err := orm.Model().Where("id",1)Delete(&t)
	// // orm.Create()
	// orm.Model().Where().Find()

	// builder.NewConnect().Connect()

	// s.Table("tbl1")
	// s.Insert(&struct {
	// 	Name string `db:"_name" db2:"test2"`
	// 	Age  int    `db:"age"`
	// }{"test", 16})
	// s.Insert(&struct {
	// 	Name string `db:"_name" db2:"test2"`
	// 	Age  int    `db:"age"`
	// }{"test", 16})
	// fmt.Println(s.BuildInsert())

	// s.Table("tbl1")
	// s.Update(&struct {
	// 	Name string `db:"_name" db2:"test2"`
	// 	Age  int    `db:"age"`
	// }{"test", 16})
	// s.Update(&struct {
	// 	Name string `db:"_name" db2:"test2"`
	// 	Age  int    `db:"age"`
	// }{"test", 16})
	// s.Where("type", "A")
	// fmt.Println(s.BuildUpdate())
	//delate
	// s.Table("tbl1.t1")
	// s.Where("type", "A")
	// s.Delete()
	// fmt.Println(s.BuildDelete())

	// c := &builder.Clause{}
	// c.Where("type", "A")
	// c.Where("status", "0")
	// c.Where(func(c *builder.Clause) {
	// 	c.Where("a", "200")
	// 	c.Where("b", "100")
	// 	c.Where(func(c *builder.Clause) {
	// 		c.Where("time", "2018")
	// 		c.Where("you", 1)
	// 	})
	// })
	// fmt.Println(c.Build(0))

	// build insert
	// m := make(map[string]interface{}, 0)
	// m["a"] = 1
	// m["b"] = 2
	// s = builder.New()
	// s.Table("tbl1")
	// s.Where("t1.status", "0")
	// s.Insert(m, m)
	// fmt.Println(s.BuildInsert())
	// fmt.Println(s.Args())
	// db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true")
	// if err != nil {
	// 	log.Println(err)
	// }
	// defer db.Close()

	var settings = make(map[string][]string, 0)
	settings["default"] = []string{
		"root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true&timeout=5s&readTimeout=6s",
		"root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true&timeout=5s&readTimeout=6s",
		"root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true&timeout=5s&readTimeout=6s",
	}
	// cluster := pool.Init("mysql", settings)
	cluster := db.InitPool("mysql", settings)
	orm.Init(cluster)
	// orm.New("de").Model()

	// m := make(map[string]interface{}, 0)
	// // m["a"] = 1
	// // m["b"] = 2
	// m["Uid"] = 2
	// s := builder.New()
	// s.Table("tbl1")
	// s.Where("t1.status", "0")
	// s.Update(m)
	// fmt.Println(s.BuildUpdate())
	// // fmt.Println(s.Args())

	t := &T{
		// Typ: &typ,
		Uid:     "1",
		Expires: 3,
	}
	rst, err := orm.Model(t).UpdateField("[+]Expires", 1).Where("id", 68).Update()
	fmt.Println("->", err, t, rst)
	fmt.Println(rst.LastInsertId())
	fmt.Println(rst.RowsAffected())

	ormx, _ := orm.Begin()
	rst, err = ormx.Model(t).UpdateField("[+]Expires", 100).Where("id", 68).Update()
	fmt.Println("->", err, t, rst)
	ormx.Rollback()
	// ormx.Commit()

	err = orm.Model(t).Where("id", 68).Fetch()
	fmt.Println("->", err, t)

	var tt []*T
	err = orm.Model(&tt).Where("id", 68).FetchAll()
	fmt.Println("->", err, tt)
	// err = orm.Model(t).Where("id", 68).Fetch()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(t)
	sql, args := builder.Select(
		builder.Table("test"),
		builder.Columns("id"),
		builder.Where("id", 68),
	)
	fmt.Println(sql, args)

	fmt.Println("--common")
	ot := &T{}
	err = db.Fetch(
		ot,
		builder.Where("id", 68),
	)
	fmt.Println("->", err, ot)

	fmt.Println("--trans begin")
	dbx, err := db.Begin()
	if err != nil {
		fmt.Println("->", err, ot)
	}
	fmt.Println("--fetch")
	dbx.Fetch(
		ot,
		builder.Where("id", 68),
	)
	fmt.Println("--update::")
	ot = &T{}
	ot.Uid = "test"
	rs, err := dbx.Update(
		ot,
		builder.Where("id", 68),
	)
	n, err := rs.RowsAffected()
	fmt.Println("->", n, err, ot)
	err = dbx.Rollback()
	fmt.Println("->>", err)
	err = dbx.Commit()
	fmt.Println("->>", err)

	//Insert
	rs, err = db.Insert(
		ot,
	)
	fmt.Println("->>", err)

	//Replace
	rs, err = db.Replace(
		ot,
	)
	fmt.Println("->>", err)

	//Delete
	rs, err = db.Delete(
		ot,
		builder.Where("1 != 1"),
	)
	fmt.Println("->>", err)

	// sql, args = builder.Update(
	// 	builder.Table("test"),
	// 	builder.Set("a", "1"),
	// 	builder.Set("b", "2"),
	// 	builder.Where("id", 68),
	// )
	// fmt.Println(sql, args)

	// ots := []T{}
	// orm.FetchAll(
	// 	ots,
	// 	orm.Where("id", 1),
	// 	orm.Limit(10),
	// 	orm.Offset(10),
	// )
	// orm.Insert(t)
	// orm.Update(t, orm.Where("id", 1))
	// orm.Delete(t, orm.Where("id", 1))

	// var typ = "11"
	// t := &T{
	// 	Typ:       &typ,
	// 	Scope:     "test",
	// 	CreatedAt: time.Now(),
	// }
	// s = builder.New()
	// s.Table("tbl1")
	// s.Where("t1.status", "0")
	// rst, err := orm.Model(t).Insert()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(rst)

}

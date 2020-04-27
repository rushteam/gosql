package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rushteam/gosql"
)

type S struct {
	Key string `db:"key"`
	Val string `db:"val"`
}
type T struct {
	ID       int64   `db:"id,pk"`
	Uid      string  `db:",index"`
	Typ      *string `db:"typ,index"`
	ClientID string  `db:"client_id"`
	Token    string  `db:"token"`
	Expires  int     `db:"expires"`
	XX       int     `db:"-"`
	Scope    string  `db:"scope,csv"`
	// Scope     json.RawMessage `db:"scope,csv"`
	UpdatedAt *time.Time `db:"updated_at"`
	CreatedAt time.Time  `db:"created_at"`
}

//TableName ..
func (t T) TableName() string {
	return "login"
}
func main() {

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
	ot.ClientID = "test"
	rs, err := dbx.Update(
		ot,
		builder.Where("id", 68),
	)
	n, err := rs.RowsAffected()
	fmt.Println("->>", n, err, ot)
	err = dbx.Rollback()
	fmt.Println("->>", err)
	err = dbx.Commit()
	fmt.Println("->>", err)

	//Insert
	rs, err = db.Insert(
		ot,
	)
	fmt.Println("--Insert>", err, ot)

	//Replace
	rs, err = db.Replace(
		ot,
	)
	fmt.Println("--Replace>", err, ot)

	//Delete
	rs, err = db.Delete(
		ot,
		builder.Where("1 != 1"),
	)
	fmt.Println("--Delete>", err)

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

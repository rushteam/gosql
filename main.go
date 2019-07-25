package main

import (
	"fmt"
	"time"

	"github.com/mlboy/godb/orm"
	"github.com/mlboy/godb/pool"

	// "github.com/didi/gendry/scanner"
	_ "github.com/go-sql-driver/mysql"
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
	// s := builder.New()
	// s.Flag("DISTINCT")
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
	// s.Where("[exists]", "select 1")
	// s.Where("[exists]", func(s *builder.SQLSegments) {
	// 	s.Table("tbl2.t2")
	// 	s.Where("xx", 10000)
	// })
	// s.GroupBy("id")
	// s.OrderBy("id desc", "id asc")
	// s.Limit(30)
	// s.Offset(10)
	// s.ForUpdate()
	// fmt.Println(s.BuildSelect())

	// type Accounts struct{}
	// db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth")

	// if err != nil {
	// 	log.Println(err)
	// }
	// defer db.Close()
	// err = db.Ping()
	// if err != nil {
	// 	log.Println(err)
	// }
	// sq := builder.New()
	// // sql.Field("*")
	// sq.Table("accounts")
	// // fmt.Println(sq.BuildSelect())
	// // rows, _ := db.Query(sq.BuildSelect())
	// rows, err := db.Query("SELECT * FROM `accounts`")
	// if err != nil {
	// 	log.Println(err)
	// }
	// var accts []Accounts
	// // fmt.Println(rows == nil)
	// err = scanner.Scan(rows, &accts)
	// if err != nil {
	// 	log.Println(err)
	// }
	// for _, acc := range accts {
	// 	fmt.Println(acc)
	// }

	//id type client_id client_secret salt created updated metadata

	// db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true")
	// if err != nil {
	// 	log.Println(err)
	// }
	// defer db.Close()
	// t := &T{}
	// rows, err := db.Query("SELECT * FROM `login` order by id desc")
	// if err != nil {
	// 	log.Println(err)
	// }
	// err = scanner.Scan(rows, t)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(t)
	// var tt []*T
	// rows, err := db.Query("SELECT * FROM `login`")
	// err = scanner.ScanAll(rows, &tt)
	// if err != nil {
	// 	log.Println(err)
	// }
	// for _, v := range tt {
	// 	fmt.Println(v)
	// }
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

	//mysql-master-def
	var settings = make(map[string]map[string][]string, 0)
	settings["default"] = make(map[string][]string, 0)
	settings["default"]["master"] = []string{"root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true"}
	cluster := pool.Init("mysql", settings)
	orm.InitCluster(cluster)

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
	// typ := "phone"
	t := &T{
		// Typ: &typ,
		Uid:     "1",
		Expires: 3,
	}
	rst, err := orm.Model(t).UpdateField("[+]Expires", 1).Where("id", 68).Update()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t)
	fmt.Println(rst.LastInsertId())
	fmt.Println(rst.RowsAffected())
	// orm.Model(t).Fetch()

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

package main

import (
	"fmt"
	"reflect"

	"./builder"
	"./scanner"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	s := builder.New()
	s.Flag("DISTANCE")
	s.Field("*")
	s.Table("tbl1.t1")
	s.Where("t1.status", "0")
	s.Where("type", "A")
	s.Where("[in]sts", []string{"1", "2", "3", "4"})
	s.Where("[in]sts2", 1)
	s.Where(func(s *builder.Clause) {
		s.Where("a", "200")
		s.Where("b", "100")
	})
	s.Where("aaa = 999")
	s.Where("[#]ccc = ?", 888)
	s.Join("tbl3", "a", "=", "b")
	s.Having("ss", "1")
	s.Where("[~]a", "AA")
	s.Where("[exists]", "AA")
	s.Where("[exists]", func(s *builder.SQLSegments) {
		s.Where("xx", 10000)
	})
	s.GroupBy("id")
	s.OrderBy("id desc", "id asc")
	s.Limit(30)
	s.Offset(10)
	s.ForUpdate()
	fmt.Println(s.BuildSelect())

	type TTT struct {
		ID   int `db:"id,pk"`
		Name int `db:"name,index"`
	}
	t := &TTT{}
	column, err := scanner.GetColumns(reflect.TypeOf(t))
	fmt.Println(err)
	fmt.Println(column)
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

	// c := &db.Clause{}
	// c.Where("type", "A")
	// c.Where("status", "0")
	// c.Where(func(c *db.Clause) {
	// 	c.Where("a", "200")
	// 	c.Where("b", "100")
	// 	c.Where(func(c *db.Clause) {
	// 		c.Where("time", "2018")
	// 		c.Where("you", 1)
	// 	})
	// })
	// fmt.Println(c.Build(0))
}

package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rushteam/gosql"
)

type UserModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (u *UserModel) TableName() string {
	return "my_user"
}

func main() {
	db := gosql.NewCluster(
		gosql.AddDb("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
	)
	user := &UserModel{}
	err := db.Fetch(user, gosql.Where("id", 1), gosql.Where("[like]name", "j%"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user)
	err := db.Fetch(user, 
		gosql.Columns("id","name"),
		gosql.Where("id", 1),
		gosql.Where("[like]name", "j%")
		gosql.OrWhere(func(s *Clause) {
			s.Where("[>=]score", "90")
			s.Where("[<=]score", "100")
		}),
		GroupBy("type"),
		OrderBy("score DESC"),
	)
	var userList []UserModel
	err := db.FetchAll(&userList, 
		gosql.Columns("id","name"),
		gosql.Where("id", 1),
		gosql.Where("[like]name", "j%")
		gosql.OrWhere(func(s *Clause) {
			s.Where("[>]score", "90")
			s.Where("[<]score", "100")
		}),
		GroupBy("type"),
		OrderBy("score DESC"),
		Offset(0),
		Limit(10),
	)
}

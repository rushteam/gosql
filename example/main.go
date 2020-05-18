package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rushteam/gosql"
)

//UserModel user model
type UserModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

//TableName for set table name
func (u *UserModel) TableName() string {
	return "my_user"
}

func main() {
	gosql.Debug = true
	var err error
	var ret gosql.Result

	db := gosql.NewCluster(
		gosql.AddDb("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
	)
	//demo 1
	user := &UserModel{}
	err = db.Fetch(user,
		gosql.Columns("id", "name"),
		gosql.Where("id", 1),
		gosql.Where("[like]name", "j%"),
		gosql.OrWhere(func(s *gosql.Clause) {
			s.Where("[>=]score", "90")
			s.Where("[<=]score", "100")
		}),
		gosql.GroupBy("type"),
		gosql.OrderBy("score DESC"),
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user)

	//demo 2
	var userList []UserModel
	err = db.FetchAll(&userList,
		gosql.Columns("id", "name"),
		gosql.Where("id", 1),
		gosql.Where("[like]name", "j%"),
		gosql.OrWhere(func(s *gosql.Clause) {
			s.Where("[>]score", "90")
			s.Where("[<]score", "100")
		}),
		gosql.GroupBy("type"),
		gosql.OrderBy("score DESC"),
		gosql.Offset(0),
		gosql.Limit(10),
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(userList)

	//demo 3
	u3 := UserModel{}
	ret, err = db.Insert(u3)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)

	//demo 5
	users := []UserModel{}
	u1 := UserModel{Name: "jack"}
	u2 := UserModel{Name: "Tom"}
	users = append(users, u1)
	users = append(users, u2)
	ret, err = db.Insert(users)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)
}

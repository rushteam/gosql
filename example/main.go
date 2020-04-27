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
		gosql.AddDb("mysql", "root:dream@tcp(127.0.0.1:3306)/rushteam?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
	)
	user := &UserModel{}
	err := db.Fetch(user, gosql.Where("id", 1), gosql.Where("[like]name", "j%"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user)
}

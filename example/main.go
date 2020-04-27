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
	fmt.Println("--common")
	ot := &T{}
	err = db.Fetch(
		ot,
		gosql.Where("id", 68),
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
		gosql.Where("id", 68),
	)
	fmt.Println("--update::")
	ot = &T{}
	ot.Uid = "test"
	ot.ClientID = "test"
	rs, err := dbx.Update(
		ot,
		gosql.Where("id", 68),
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
		gosql.Where("1 != 1"),
	)
	fmt.Println("--Delete>", err)
}

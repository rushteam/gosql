package orm

import (
	"database/sql"
	"testing"
)

func TestModel(t *testing.T) {
	t.Skip()
}
func TestFind(t *testing.T) {
	type S struct {
		ID  string `db:"id"`
		Typ string `db:"type"`
	}
	s := &S{}
	db, err := sql.Open("mysql", "root:123321@tcp(127.0.0.1:3306)/auth?parseTime=true")
	if err != nil {
		panic(err)
	}
	orm := Model(db).Find(s)
	t.Log(orm)

}

package gosql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type t1Model struct {
	Name string `db:"name"`
}

func (t *t1Model) TableName() string {
	return "test"
}

type t2Model struct {
	ID   int64  `db:"id,pk"`
	Name string `db:"name"`
}

func (t *t2Model) TableName() string {
	return "test"
}
func TestError(t *testing.T) {
	Debug = true
	err := &Error{10062, "test"}
	t.Log(err)
	t.Log(err.Error())
}

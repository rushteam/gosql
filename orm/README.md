# ORM Doc

## 引用

import "github.com/mlboy/godb/orm"

## struct
type UserModel struct {
    Username `db:"username"`
    Num `db:"-"`
    CreatedAt    time.Time `db:"created_at"`
}
//TableName ..
func (UserModel) TableName() string {
	return "pay_trade"
}
## insert
m := &UserModel{}
rst, err := orm.Model(m).Insert()

## update
m := &UserModel{}
rst, err := orm.Model(m).Update()


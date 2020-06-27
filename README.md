# gosql

[![GoTest](https://github.com/rushteam/gosql/workflows/Go/badge.svg)](https://github.com/rushteam/gosql/actions)
[![GoDoc](https://godoc.org/github.com/rushteam/gosql?status.svg)](https://pkg.go.dev/github.com/rushteam/gosql)
[![codecov](https://codecov.io/gh/rushteam/gosql/branch/master/graph/badge.svg)](https://codecov.io/gh/rushteam/gosql)
[![Go Report Card](https://goreportcard.com/badge/github.com/rushteam/gosql)](https://goreportcard.com/report/github.com/rushteam/gosql)
[![LICENSE](https://img.shields.io/github/license/rushteam/gosql)](https://github.com/rushteam/gosql/blob/master/LICENSE)

gosql is a easy ORM library for Golang. 

## Feature

* Golang-style SQL builder
* Unlimited nesting query
* Reading and Writing Separation
* Delay connection creation
* ORM maping to sturct
* Support transaction
* Versatile
* Clean Code
* Bulk Insert

## Structure

* db.go: Basic struct definition
* pool.go: Manage DB pool
* session.go: Session and Model
* builder.go: Building SQL
* scanner/*: scan struct

## Why build this wheels

I have read almost all open source operation database library implemented in golang on github. But never get the optimal solution.

Such as these:

1. gorm: Does not support read and write separation.

2. gendry: Occupy special keywords and partially ugly syntax.

3. sqlx: Mostly good, But the syntax is not simple enough, and does not support the separation of reading and writing.

This project refers to a large number of existing libs, refers to various documents, and uses golang style to achieve from scratch.

## NOTE

NOTE: Only supports mysql driver.

## Demo

Let's look a demo frist.

```sql
SELECT DISTINCT *
FROM `tbl1`.`t1`
    JOIN `tbl3` ON `a` = `b`
WHERE (`t1`.`status` = ?
    AND `name` = ?
    AND `nick` != ?
    AND `role1` IN (?, ?, ?, ?)
    AND `role2` NOT IN (?, ?, ?, ?)
    AND `card1` IN (?)
    AND `card2` NOT IN (?)
    AND (`age` > ?
        AND `age` < ?)
    AND v1 = 1
    AND v2 = ?
    AND `desc` LIKE ?
    AND `desc` NOT LIKE ?
    AND EXISTS (
        SELECT 1
    )
    AND NOT EXISTS (
        SELECT *
        FROM `tbl2`.`t2`
        WHERE `t2`.`id` = ?
    ))
GROUP BY `class,group`
HAVING `class` = ?
ORDER BY `score desc`, `name asc`
LIMIT 10, 30
FOR UPDATE
```

```golang
    s := gosql.NewSQLSegment()
    s.Flag("DISTINCT")
    s.Field("*")
    s.Table("tbl1.t1")
    s.Where("t1.status", "0")
    s.Where("name", "jack")
    s.Where("[!=]nick", "tom")
    s.Where("[in]role1", []string{"1", "2", "3", "4"})
    s.Where("[!in]role2", []string{"1", "2", "3", "4"})
    s.Where("[in]card1", 1)
    s.Where("[!in]card2", 1)
    s.Where(func(s *Clause) {
        s.Where("[>]age", "20")
        s.Where("[<]", "50")
    })
    s.Where("v1 = 1")
    s.Where("[#]v2 = ?", 2)
    s.Join("tbl3", "a", "=", "b")
    s.Having("class", "one")
    s.Where("[~]desc", "student")
    s.Where("[!~]desc", "teacher")
    s.Where("[exists]my_card", "select 1")
    s.Where("[!exists]my_card2", func(s *SQLSegments) {
        s.Table("tbl2.t2")
        s.Where("t2.id", 10000)
    })
    s.GroupBy("class,group")
    s.OrderBy("score desc", "name asc")
    s.Limit(30)
    s.Offset(10)
    s.ForUpdate()
    fmt.Println(s.BuildSelect())
```

## Getting Started

```golang
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
}

```

## Doc

### Debug Mode

```golang
//this code will be start at debug mode and the sql will be print
gosql.Debug = true
```

### Struct Model

To define a Model struct, use the struct and tag syntax.

#### Simple define a model

```golang
type User struct {
    ID int64
    Age int
    Name string
    CreatedAt time.Time
}
```

Usually define a Struct can be used as a model, gosql will parse out the table name, field mapping relationship,etc.

table: user
columns: id,age,name,created_at

#### Using tag syntax

Use structure tags to customize field mapping

```golang
type User struct {
    ID int64 `db:"uid,pk"`
    Age int `db:"age"`
    Name string `db:"fisrt_name"`
    CreatedAt time.Time `db:"created_at"`
}
```

table: user
columns: uid,age,fisrt_name,created_at
pk: uid

#### Define table name

Implement "TableName" method to specify the table name

```golang
type User struct {}
func (u *User) TableName() string {
    return "my_user"
}
```

table: my_user

### Exec

#### INSERT

db.Insert(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
user.Name = "jack"
ret,err := db.Insert(&user)
```

batch insert

```golang
users := []UserModel{}
u1 := UserModel{Name:"jack"}
u2 := UserModel{Name:"Tom"}
users = append(users,u1)
users = append(users,u2)
ret,err := db.Insert(users)
```

#### REPLACE

db.Replace(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
user.Name = "jack"
ret,err := db.Replace(&user,gosql.Where("id",1))
```

#### UPDATE

Update(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
user.Name = "jack Ma"
ret,err := db.Update(&user,gosql.Where("id",1))
```

#### DELETE

db.Delete(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
ret,err := db.Delete(&user,gosql.Where("id",1))
//sql: delete from my_user where id = 1
```

### QUERY

#### Get a record: db.Fetch(dst interface{}, opts ...Option) error

```golang
user := &UserModel{}
err := db.Fetch(user,
    gosql.Columns("id","name"),
    gosql.Where("id", 1),
    gosql.Where("[like]name", "j%"),
    gosql.OrWhere(func(s *Clause) {
        s.Where("[>=]score", "90")
        s.Where("[<=]age", "100")
    }),
    gosql.GroupBy("type"),
    gosql.OrderBy("score DESC"),
)
```

#### Get multiple records: db.FetchAll(dst interface{}, opts ...Option) error

```golang
var userList []UserModel
err := db.FetchAll(&userList,
    gosql.Columns("id","name"),
    gosql.Where("id", 1),
    gosql.Where("[like]name", "j%"),
    gosql.OrWhere(func(s *Clause) {
        s.Where("[>]score", "90")
        s.Where("[<]score", "100")
    }),
    gosql.GroupBy("type"),
    gosql.OrderBy("score DESC"),
    gosql.Offset(0),
    gosql.Limit(10),
)
```

### OPTION

#### WHERE

* gosql.Where("id",1)

```golang
gosql.Where("id",1)
//sql: id = 1
```

* gosql.Where("[>]age",18)

```golang
gosql.Where("[>]age",18)
//sql: age > 18
```

* gosql.Where("[in]id",[]int{1,2})

```golang
gosql.Where("[in]id",[]int{1,2})
//sql: id in (1,2)
```

* gosql.Where("[!in]id",[]int{1,2})

```golang
gosql.Where("[!in]id",[]int{1,2})
//sql: id not in (1,2)
```

* gosql.Where("[~]name","ja%")

```golang
gosql.Where("[~]name","ja%")
//sql: name like 'ja%'
```

* gosql.Where("[!~]name","ja%")

```golang
gosql.Where("[!~]name","ja%")
//sql: name not like 'ja%'
```

#### symbol [?]

* [=] equal

```golang
gosql.Where("[=]id",1)
//sql: id = 1
```

* [!=] not equal

```golang
gosql.Where("[!=]id",1)
//sql: id != 1
```

* [>] greater than

```golang
gosql.Where("[>]id",1)
//sql: id > 1
```

* [>=] greater or equal

```golang
gosql.Where("[>=]id",1)
//sql: id >= 1
```

* [<] less

```golang
gosql.Where("[<]id",1)
//sql: id < 1
```

* [<=] less or equal

```golang
gosql.Where("[<=]id",1)
//sql: id <= 1
```

* [in] in

```golang
gosql.Where("[in]id",[]int{1,2})
//sql: id in (1,2)
```

* [!in] not in

```golang
gosql.Where("[!in]id",[]int{1,2})
//sql: id not in (1,2)
```

* [is] is null

```golang
gosql.Where("[is]name",nil)
//sql: name is null
```

* [!is] not is null

```golang
gosql.Where("[!is]name",nil)
//sql: id is not null
```

* [exists] exists

```golang
gosql.Where("[exists]name","select 1")
//sql: name exists(select 1)
```

* [!exists] not exists

```golang
gosql.Where("[!exists]name","select 1")
//sql: name not exists(select 1)
```

* [#] sql

```golang
gosql.Where("[#]age=age-1")
//sql: age = age-1
```

### Raw SQL: db.Query()

```golang
rows,err := db.Query("select * from my_user where id = ?",1)
//sql: select * from my_user where id = 1
```

### select master or slave

* db.Master() change to master

```golang
db := db.Master()
db.Fetch(...)
```

* db.Slave() change to slave

```golang
db := db.Slave()
db.Fetch(...)
```

### Paging

Define a page function and return gosql.Option sturct

```golang 
//Page  pn: per page num ,ps: page size
func Page(pn, ps int) gosql.Option {
	if pn < 1 {
		pn = 1
	}
	return func(s gosql.SQLSegments) gosql.SQLSegments {
		s.Limit(ps)
		s.Offset((pn - 1) * ps)
		return s
	}
}
func main() {
    user := &UserModel{}
    err := db.Fetch(user,
        Page(1,15),
    )
}

```

### multi-database 

```golang
gosql.NewCollect(
    gosql.NewCluster(
        gosql.AddDb("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
    ),
    "db1",
)
gosql.NewCollect(
    gosql.NewCluster(
        gosql.AddDb("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
    ),
    "db2",
)

db1 := gosql.Collect("db1")


db2 := gosql.Collect("db2")
```

### builder of API

* builder.New() start a builder

```golang
s := builder.New()
```

* builder.Flag(f string) set a flag

```golang
s.Flag("test")
```

* builder.Field(fields string) Specified columns

default value *

```golang
s.Field("*")
```

* builder.Table(tbl string) Specified table name

```golang
s.Table("tbl.t1")
```

#### Where

builder.Where(key string, val inferface{})

* Eq

```golang
s.Where("t1.status", "0")
//sql: t1.status = 0
```

* Not Eq

```golang
s.Where("[!=]t1.status", "0")
//sql: t1.status != 0
```

* In

```golang
s.Where("[in]field", []string{"a", "b", "c"})
//sql: t1.field in (a,b,c)
```

* No In

```golang
s.Where("[!in]field", []string{"a", "b", "c"})
//sql: t1.status in (a,b,c)
```

### Nested Where

* s.Where(func(s *builder.Clause){}

```golang
s.Where("[!]t1.a",1).Where(func(s *builder.Clause){
    s.Where("t1.b",1)
    s.OrWhere("t1.c",1)
})
//sql: t1.a != 1  and (t1.b = 1 or t1.c = 1)
```

### Other statements

* Group By

```golang
s.GroupBy("class")
//sql: group by `class`
```

* Order By

```golang
s.OrderBy("id desc", "age asc")
//sql: order by `id` desc, `age` asc
```

* Limit

```golang
s.Limit(10)
//sql: limit 10
```

* Offset

```golang
s.Offset(10)
//sql: offset 10
```

## Contributing

When everybody adds fuel, the flames rise high.

Let's build our self library.

You will be a member of rushteam which is An open source organization

Thanks for you, Good Lucy.

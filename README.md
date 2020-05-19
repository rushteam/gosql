# gosql

[![GoTest](https://github.com/rushteam/gosql/workflows/Go/badge.svg)](https://github.com/rushteam/gosql/actions)
[![GoDoc](https://godoc.org/github.com/rushteam/gosql?status.svg)](https://pkg.go.dev/github.com/rushteam/gosql)
[![codecov](https://codecov.io/gh/rushteam/gosql/branch/master/graph/badge.svg)](https://codecov.io/gh/rushteam/gosql)
[![Go Report Card](https://goreportcard.com/badge/github.com/rushteam/gosql)](https://goreportcard.com/report/github.com/rushteam/gosql)
[![LICENSE](https://img.shields.io/github/license/rushteam/gosql)](https://github.com/rushteam/gosql/blob/master/LICENSE)

A easy ORM for golang

gosql 是一个用golang实现的数据库操作类库

## Feature 功能

* Golang-style SQL builder go语言风格sql生成
* Unlimited nesting query 查询条件无限嵌套
* Reading and Writing Separation 读写分离
* Delay connection creation 延迟创建连接
* ORM maping to sturct ORM映射结构体
* Transactions 事务支持
* Versatile 功能多样的
* Clean Code 简洁的代码
* Bulk Insert 支持批量插入

## Structure 结构

* db.go: defined base struct define 基本结构定义
* pool.go: db manager 管理db
* session.go: session and maping to model 会话和模型
* builder.go: for building SQL 构建sql
* scanner/*: mapping struct and scan 映射模型

## Why build this wheels 为什么造轮子

几乎是翻遍github上所有开源的使用golang实现的操作数据库类)使用起来总有不顺手的地方,例如:

gorm不支持读写分离,关联表使用频率少

gendry 是didi开源的一款,比较简洁但部分语法怪异 如group by 和 limit 依赖字段的约定

sqlx 相比起来不错,但语法不够简洁,不支持读写分离,

gosql 目前仅支持mysql （关键是`符号的处理，以及一些特殊语法，后期可能会考虑兼容pgsql等

本数据库参阅了大量现有数据库架构,参阅各种文献,自定义语法习惯,从零实现

其中灵感来自:分模块的灵感来自gendry,标签读取部分参考gorm,拼装sql的语法来自于我之前写的php的操作db库

## DEMO 例子

为了展示gosql的能力,先展示个例子:
Let's look a demo:

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

## Getting Started 开始使用

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

## Doc 文档

## Debug

```golang
//this code will be start at debug mode and the sql will be print
gosql.Debug = true
```

## Struct Model

To define a model structure, just use the struct syntax.

### Simple define a model

```golang
type User struct {
    ID int64
    Age int
    Name string
    CreatedAt time.Time
}
```

Usually define a Struct can be used as a model, gosql will parse out the table name, field mapping relationship,etc.

定义一个普通的结构体就可以被作为一个model , gosql 将会从中解析出表名、字段映射关系等

table: user
columns: ID,Age,Name,CreatedAt

### Using tag syntax

Use structure tags to customize field mapping

使用结构体tags 可以自定义字段映射

```golang
type User struct {
    ID int64 `db:"uid,pk"`
    Age int `db:"age"`
    Name string `db:"nickname"`
    CreatedAt time.Time `db:"created_at"`
}
```

table: user
columns: uid,age,nickname,created_at
pk: uid

### Using custom table name

Implement "TabbleName" method to specify the table name

通过实现 TabbleName 方法来指定一个表名

```golang
type User struct {
    ID int64 `db:"uid,pk"`
    Age int `db:"age"`
    Name string `db:"nickname"`
    CreatedAt time.Time `db:"created_at"`
}
func (u *User) TableName() string {
    return "my_user"
}
```

table: my_user

## Auto

## Exec

### INSERT

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

### REPALCE

db.Replace(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
user.Name = "jack"
ret,err := db.Replace(&user,gosql.Where("id",1))
```

### UPDATE

Update(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
user.Name = "jack Ma"
ret,err := db.Update(&user,gosql.Where("id",1))
```

### DELETE

db.Delete(dst interface{}, opts ...Option) (Result, error)

```golang
user := &UserModel{}
ret,err := db.Delete(&user,gosql.Where("id",1))
//sql: delete from my_user where id = 1
```

## QUERY

### Get a record: db.Fetch(dst interface{}, opts ...Option) error

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

### Get multiple records: db.FetchAll(dst interface{}, opts ...Option) error

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

## OPTION

### WHERE

#### gosql.Where("id",1)

```golang
gosql.Where("id",1)
//sql: id = 1
```

#### gosql.Where("[>]age",18)

```golang
gosql.Where("[>]age",18)
//sql: age > 18
```

#### gosql.Where("[in]id",[]int{1,2})

```golang
gosql.Where("[in]id",[]int{1,2})
//sql: id in (1,2)
```

#### gosql.Where("[!in]id",[]int{1,2})

```golang
gosql.Where("[!in]id",[]int{1,2})
//sql: id not in (1,2)
```

#### gosql.Where("[~]name","ja%")

```golang
gosql.Where("[~]name","ja%")
//sql: name like 'ja%'
```

#### gosql.Where("[!~]name","ja%")

```golang
gosql.Where("[!~]name","ja%")
//sql: name not like 'ja%'
```

### 条件表达式 [?]

#### [=] equal

```golang
gosql.Where("[=]id",1)
//sql: id = 1
```

#### [!=] not equal

```golang
gosql.Where("[!=]id",1)
//sql: id != 1
```

#### [>] greater than

```golang
gosql.Where("[>]id",1)
//sql: id > 1
```

#### [>=] greater or equal

```golang
gosql.Where("[>=]id",1)
//sql: id >= 1
```

#### [<] less

```golang
gosql.Where("[<]id",1)
//sql: id < 1
```

#### [<=] less or equal

```golang
gosql.Where("[<=]id",1)
//sql: id <= 1
```

#### [in] in

```golang
gosql.Where("[in]id",[]int{1,2})
//sql: id in (1,2)
```

#### [!in] not in

```golang
gosql.Where("[!in]id",[]int{1,2})
//sql: id not in (1,2)
```

#### [is] is null

```golang
gosql.Where("[is]name",nil)
//sql: name is null
```

#### [!is] not is null

```golang
gosql.Where("[!is]name","")
//sql: id is not ""
```

#### [exists] exists

```golang
gosql.Where("[exists]name","select 1")
//sql: name exists(select 1)
```

#### [!exists] not exists

```golang
gosql.Where("[!exists]name","select 1")
//sql: name not exists(select 1)
```

#### [#] sql

```golang
gosql.Where("[#]age=age-1")
//sql: age = age-1
```

## Raw SQL: db.Query()

```golang
rows,err := db.Query("select * from my_user where id = ?",1)
//sql: select * from my_user where id = 1
```

## select master or slave

### change to master: db.Master()

```golang
db := db.Master()
db.Fetch(...)
```

### change to slave: db.Slave()

```golang
db := db.Slave()
db.Fetch(...)
```

## builder of API

### 创建语句

**用法** builder.New()

例子 s := builder.New()

### 设置Flag builder.Flag(f string)

设置一个falg,非必须

**用法** s.Flag(string)

例子 s := builder.New().Flag("")

### 指定字段 builder.Field(fields string)

指定查询字段 不指定 默认为 *

**用法** s.Field("*")

### 指定表名 builder.Table(tbl string)

**用法** s.Table("tbl1.t1")

### 查询条件

* 普通查询 s.Where(key string, val inferface{})

* 等于查询

 **用法** s.Where("t1.status", "0")

 **等效SQL** t1.status = 0

* 不等于查询

**用法** s.Where("[!=]t1.status", "0")

**等效SQL** t1.status != 0

* IN查询

**用法** s.Where("[in]sts", []string{"a", "b", "c"})

**等效SQL** t1.type in (a,b,c)

* NOT IN查询

**用法** s.Where("[!in]sts", []string{"a", "b", "c"})

**等效SQL** t1.type not in (a,b,c)

* 复杂条件查询

**用法** .Where(func(s *builder.Clause){}

```golang
s.Where("[!]t1.a",1).Where(func(s *builder.Clause){
    s.Where("t1.b",1)
    s.OrWhere("t1.c",1)
})
```

**等效SQL**  t1.a != 1  and (t1.b = 1 or t1.c = 1)

* GROUP BY 分类

**用法**  s.GroupBy("id")

**等效SQL** group by `id`

* ORDER BY 排序

**用法**  s.OrderBy("id desc", "age asc")

**等效SQL** order by `id` desc

* 限制条数

**用法**  s.Limit(30)

**等效SQL** limit 30

* 偏移条数

**用法**  s.Offset(10)

**等效SQL** offset 30

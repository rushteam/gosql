# godb

A yet ORM for golang

gosql 是一个数据库的golang库

## Why build this wheels

几乎是翻遍github上所有开源的使用golang实现的操作数据库类)使用起来总有不顺手的地方,例如:
gorm不支持读写分离,关联表使用频率少
gendry 是didi开源的一款,比较简洁但部分语法怪异 如group by 和 limit 依赖字段的约定
sqlx 相比起来不错,但语法不够简洁,不支持读写分离,

gosql 目前仅支持mysql （关键是`符号的处理，以及一些特殊语法，后期可能会考虑兼容pgsql等
本数据库参阅了大量现有数据库架构,参阅各种文献,自定义语法习惯,从零实现
其中灵感来自:分模块的灵感来自gendry,标签读取部分参考gorm,拼装sql的语法来自于我之前写的php的操作db库

## Structure

* db.go: 基本结构定义
* pool.go: 实现主从链接管理
* session.go: 实现会话,model映射
* builder.go: 主要负责拼接SQL for building SQL
* scanner/: 映射数据到结构体 mapping struct and scan


## Feature

* Versatile 功能多样的
* Unlimited nesting query 查询条件无限嵌套
* Reading and Writing Separation 读写分离

## Builder of DEMO

先举个例子说明gosql的能力
下面这条复杂的sql用,如何用编写？

```sql
SELECT DISTINCT * FROM `tbl1`.`t1` JOIN `tbl3` ON `a` = `b`
WHERE (`t1`.`status` = ?
  AND `type` = ?
  AND `sts` IN (?, ?, ?, ?)
  AND `sts2` IN (?)
  AND (`a` = ? AND `b` = ?)
  AND aaa = 999
  AND ccc = ?
  AND `a` LIKE ? AND EXISTS (SELECT 1)
  AND EXISTS (SELECT * FROM `tbl2`.`t2` WHERE `xx` = ?)
) GROUP BY `id` HAVING `ss` = ? ORDER BY `id desc`, `id asc` LIMIT 10, 30
FOR UPDATE
```

```golang
    s := builder.New()
    s.Flag("DISTINCT")
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
    s.Where("[exists]", "select 1")
    s.Where("[exists]", func(s *builder.SQLSegments) {
        s.Table("tbl2.t2")
        s.Where("xx", 10000)
    })
    s.GroupBy("id")
    s.OrderBy("id desc", "id asc")
    s.Limit(30)
    s.Offset(10)
    s.ForUpdate()
    fmt.Println(s.BuildSelect())
```

## How to use

1. 初始化DB Init db

```golang
db,err := gosql.NewCluster(
    gosql.AddDb("mysql","user:pasword@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s"),
)
if err != nil {
    fmt.Println(err)
}
```


## 文档 Doc

## 操作 exec
### 插入 db.Insert(dst interface{}, opts ...Option) (Result, error)
```
type UserModel struct{
    ID int `db:"id"`
    Name string ``
}
func (u *UserModel) TableName() {
    return "my_world"
}
user := &WorldModel{}
user.Name = "jack"
ret,err := db.Insert(&user)

```
### 替换 db.Replace(dst interface{}, opts ...Option) (Result, error)
```
type UserModel struct{
    ID int `db:"id"`
    Name string ``
}
func (u *UserModel) TableName() {
    return "my_world"
}
user := &WorldModel{}
user.Name = "jack"
ret,err := db.Replace(&user,gosql.Where("id",1))

```
### 修改 Update(dst interface{}, opts ...Option) (Result, error)
### 删除 db.Delete(dst interface{}, opts ...Option) (Result, error)

## 查询 query

### 查询一条记录 db.Fetch(dst interface{}, opts ...Option) error

### 查询多条记录 db.FetchAll(dst interface{}, opts ...Option) error

## 选项 option

### 条件 gosql.Where()

#### gosql.Where("id",1)
eq sql:
```sql
id = 1
```
#### gosql.Where("age[>]",18)
eq sql:
```sql
age > 18
```
#### gosql.Where("id[in]",[]int{1,2})
eq sql:
```sql
id in (1,2)
```
#### gosql.Where("id[!in]",[]int{1,2})
eq sql:
```sql
id not in (1,2)
```
#### gosql.Where("name[~]","ja%")

eq sql:
```sql
name like 'ja%'
```

#### gosql.Where("name[!~]","ja%")

eq sql:
```sql
name not like 'ja%'
```

### 条件表达式 [?]

#### [=] 等于 
```
gosql.Where("[=]id",1)
//sql: id = 1
```
#### [!=] 不等于 
```
gosql.Where("[!=]id",1)
//sql: id != 1
```
#### [>] 大于
```
gosql.Where("[>]id",1)
//sql: id > 1
```
#### [>=] 大于等于
```
gosql.Where("[>=]id",1)
//sql: id >= 1
```

#### [<] 小于
```
gosql.Where("[<]id",1)
//sql: id < 1
```
#### [<=] 小于等于
```
gosql.Where("[<=]id",1)
//sql: id <= 1
```

#### [in] in()
```
gosql.Where("[in]id",[]int{1,2})
//sql: id in (1,2)
```

#### [!in] not in()
```
gosql.Where("[!in]id",[]int{1,2})
//sql: id not in (1,2)
```

#### [is] is
```
gosql.Where("[is]name",nil)
//sql: name is null
```

#### [!is] not is
```
gosql.Where("[!is]name","")
//sql: id is not ""
```

#### [exists] not is
```
gosql.Where("[exists]name","select 1")
//sql: name exists(select 1)
```
#### [!exists] not is
```
gosql.Where("[!exists]name","select 1")
//sql: name not exists(select 1)
```

#### [#] sql
```
gosql.Where("[#]age=age-1")
//sql: age = age-1 
```

## 原生SQL db.Query()
```
rows,err := db.Query("select * from world where id = ?",1)
```

## 主从选择

### 强制选择主 db.Master()
```
master := db.Master()
master.Fetch()
```

### 强制选择从 db.Slave()

```
master := db.Slave()
master.Fetch()
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
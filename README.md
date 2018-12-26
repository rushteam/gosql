# godb

godb 是一个链式操作数据库的golang库

已经有那么多操作db的库了，为什么还要写godb？

因为市面上所有的db库用起来总有不顺手的地方,比如gorm不支持读写分离,关联表使用频率少,比如sqlx语法不够简洁,比如gendry group by 、limit 语法怪异

godb 是分模块化的一个db操作库 目前仅支持mysql （关键是`符号的处理，以及一些特殊语法，后期可能会考虑兼容pgsql）分模块的灵感来自gendry,标签读取部分参考gorm,拼装sql的语法来自于我之前写的php的操作db库

## structure:

* builder 拼装sql
* scanner 映射数据到结构体
* orm
* mannger 数据库管理（读写分离）

## feature:

* 链式操作
* 查询条件无限嵌套
* 读写分离
* 数据库连接池


## builder of DEMO:

先看看这条复杂的sql用builder如何实现？

```sql

SELECT DISTINCT *
FROM `tbl1`.`t1`
	JOIN `tbl3` ON `a` = `b`
WHERE (`t1`.`status` = ?
	AND `type` = ?
	AND `sts` IN (?, ?, ?, ?)
	AND `sts2` IN (?)
	AND (`a` = ?
		AND `b` = ?)
	AND aaa = 999
	AND ccc = ?
	AND `a` LIKE ?
	AND EXISTS (
		SELECT 1
	)
	AND EXISTS (
		SELECT *
		FROM `tbl2`.`t2`
		WHERE `xx` = ?
	))
GROUP BY `id`
HAVING `ss` = ?
ORDER BY `id desc`, `id asc`
LIMIT 10, 30
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

## builder of API:

### 创建语句

用法 builder.New()

例子 s := builder.New()


### 设置Flag builder.Flag(f string)

设置一个falg,非必须

用法 s.Flag(string)

例子 s := builder.New().Flag("")


### 指定字段 builder.Field(fields string)

指定查询字段 不指定 默认为 *

用法 s.Field("*")

### 指定表名 builder.Table(tbl string)

用法 s.Table("tbl1.t1")

### 查询条件 

* 普通查询 s.Where(key string, val inferface{})

 * 等于查询 

    用法 s.Where("t1.status", "0")

    等效SQL t1.status = 0

  * 不等于查询 
 
    用法 s.Where("[!]t1.status", "0")

    等效SQL t1.status != 0

* IN查询

    用法 s.Where("[in]sts", []string{"a", "b", "c"})

    等效SQL t1.type in (a,b,c)

* NOT IN查询

    用法 s.Where("[!in]sts", []string{"a", "b", "c"})

    等效SQL t1.type not in (a,b,c)

* 复杂条件查询

    用法

    ```golang
    s.Where("[!]t1.a",1).Where(func(s *builder.Clause){
        s.Where("t1.b",1)
        s.OrWhere("t1.c",1)
    })
    ```

    等效SQL  t1.a != 1  and (t1.b = 1 or t1.c = 1)
    
* GROUP BY 分类

    用法  s.GroupBy("id")

    等效SQL group by `id`

* ORDER BY 排序

    用法  s.OrderBy("id desc", "age asc")

    等效SQL order by `id` desc

* 限制条数

    用法  s.Limit(30)

    等效SQL limit 30

* 偏移条数

    用法  s.Offset(10)

    等效SQL offset 30
	






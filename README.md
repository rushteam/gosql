# godb

godb 是一个链式操作数据库的golang库

已经有那么多操作db的库了，为什么还要写godb？

因为市面上所有的db库用起来总有不顺手的地方
比如gorm不支持读写分离,关联表使用频率少
比如sqlx语法不够简洁
比如gendry group by 、limit 语法怪异

godb 是分模块化的一个db操作库 目前仅支持mysql （关键是`符号的处理，以及一些特殊语法，后期可能会考虑兼容pgsql）
分模块的灵感来自gendry,标签读取部分参考gorm,拼装sql的语法来自于我之前写的php的操作db库

结构:
    builder 拼装sql
    scanner 映射数据到结构体
    orm
    mannger 数据库管理（读写分离）

feature:
    链式操作
    查询条件无限嵌套
    读写分离
    数据库连接池


builder of DEMO:

先看看这条复杂的sql用builder如何实现？
    s := builder.New()
	s.Flag("DISTANCE")
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
	s.Where("[exists]", "AA")
	s.Where("[exists]", func(s *builder.SQLSegments) {
		s.Where("xx", 10000)
	})
	s.GroupBy("id")
	s.OrderBy("id desc", "id asc")
	s.Limit(30)
	s.Offset(10)
	s.ForUpdate()
	// fmt.Println(s.BuildSelect())

builder of API:
    
    新开始一条语句
    s := builder.New()

    设置一个falg,非必须 
    s.Flag(string)

    指定字段
    s.Field("*")

    指定表名
    s.Table("tbl1.t1")

    查询条件 t1.status = 0
    s.Where("t1.status", "0")

    查询条件  t1.type in (a,b,c)
    s.Where("[in]sts", []string{"a", "b", "c"})

    查询条件  t1.a != 1  and (t1.b = 1 or t1.c = 1)

    s.Where("[!]t1.a",1).Where(func(s *builder.Clause){
        s.Where("t1.b",1)
        s.OrWhere("t1.c",1)
    })

    分类
    s.GroupBy("id")
    排序
    s.OrderBy("id desc", "id asc")
    条数
    s.Limit(30)
    位移
	s.Offset(10)






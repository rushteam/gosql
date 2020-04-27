还没写例子,这里全是测试的代码请忽略
```
    db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	t := &T{}
	rows, err := db.Query("SELECT * FROM `login` order by id desc")
	if err != nil {
		log.Println(err)
	}
	err = scanner.Scan(rows, t)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(t)
	var tt []*T
	rows, err := db.Query("SELECT * FROM `login`")
	err = scanner.ScanAll(rows, &tt)
	if err != nil {
		log.Println(err)
	}
	for _, v := range tt {
		fmt.Println(v)
	}
```
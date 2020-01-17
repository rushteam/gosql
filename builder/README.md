# sql builder

## style with options func

### select

select * from test where id = 68

```
sql, args := builder.Select(
    builder.Table("test"),
    builder.Columns("id"),
    builder.Where("id", 68),
)
fmt.Println(sql, args)
```

Use only builder

```
// type Accounts struct{}
	// db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth")

	// if err != nil {
	// 	log.Println(err)
	// }
	// defer db.Close()
	// err = db.Ping()
	// if err != nil {
	// 	log.Println(err)
	// }
	// sq := builder.New()
	// // sql.Field("*")
	// sq.Table("accounts")
	// // fmt.Println(sq.BuildSelect())
	// // rows, _ := db.Query(sq.BuildSelect())
	// rows, err := db.Query("SELECT * FROM `accounts`")
	// if err != nil {
	// 	log.Println(err)
	// }
	// var accts []Accounts
	// // fmt.Println(rows == nil)
	// err = scanner.Scan(rows, &accts)
	// if err != nil {
	// 	log.Println(err)
	// }
	// for _, acc := range accts {
	// 	fmt.Println(acc)
	// }

```

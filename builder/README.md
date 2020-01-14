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

package scanner

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

/**
tag 用法
column 设置行
	`db:"column:'id'"`
pk 设为主键 (primary_key)
index 普通索引
unique_index 唯一索引
auto 自增 (AUTO_INCREMENT)
size 字段长度
- 忽略字段
type 字段类型
	type:varchar(100)
其他
	not null;unique
**/
const tagKey = "db"
const tagSplit = ","

type structField struct {
	column     string
	index      int
	primaryKey bool
}
type structData struct {
	columns []string
	fields  map[string]*structField
	pk      string
}

//反射结构体缓存
var refStructCache = make(map[reflect.Type]*structData)
var refStructCacheMutex sync.Mutex

func GetColumns(dstType reflect.Type) (*structData, error) {
	refStructCacheMutex.Lock()
	defer refStructCacheMutex.Unlock()

	if result, ok := refStructCache[dstType]; ok {
		return result, nil
	}

	if dstType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("scanner called with non-pointer destination %v", dstType)
	}
	structType := dstType.Elem()
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("scanner called with pointer to non-struct %v", dstType)
	}

	data := new(structData)
	data.fields = make(map[string]*structField)

	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		// skip non-exported fields
		if f.PkgPath != "" {
			continue
		}
		tags := strings.Split(f.Tag.Get(tagKey), tagSplit)
		// skip using "-" tag fields
		if len(tags) > 0 && tags[0] == "-" {
			continue
		}
		// default to the field name
		name := f.Name
		// just a name in tag
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		} else {
			//大小写转换下划线的、自定义方法的
			// name = Mapper(f.Name)
		}
		// check for a meddler
		// var meddler Meddler = registry["identity"]
		sf := &structField{
			column:     name,
			primaryKey: name == data.pk,
			// index:      idx,
			// meddler:    meddler,
		}
		for i, tag := range tags {
			t := strings.Split(tag, ":")
			if len(t) < 1 {
				t = append(t, tag)
			}
			switch strings.ToLower(t[0]) {
			case "primary key", "pk", "primary", "primary_key": //primary_key
				//pk can not is a pointer
				if f.Type.Kind() == reflect.Ptr {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key but is a pointer", f.Name)
				}
				//primary key can only be one
				if data.pk != "" {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key, but a primary key field was already found", f.Name)
				}
				data.pk = name
				// } else if m, ok := registry[tag]; ok {
				// 	// meddler = m
				// } else {
				// 	return nil, fmt.Errorf("scanner found field %s with meddler %s, but that meddler is not registered", f.Name, tag)
			case "unique_index", "unique": //unique_index

			case "index": //index

			}
			sf.index = i
		}
		if _, ok := data.fields[name]; ok {
			return nil, fmt.Errorf("scanner found multiple fields for column %s", name)
		}
		sf.column = name
		sf.primaryKey = name == data.pk
		data.fields[name] = sf
		data.columns = append(data.columns, name)
	}
	refStructCache[dstType] = data
	return data, nil
}

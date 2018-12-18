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
const tagSplit = ";"

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

func getColumns(dstType reflect.Type) (*structData, error) {
	refStructCacheMutex.Lock()
	defer refStructCacheMutex.Unlock()

	if result, present := refStructCache[dstType]; present {
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

		// the tag can override the field name
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		} else {
			// use mapper func if field has no explicit tag
			// name = Mapper(f.Name)
		}

		// check for a meddler
		// var meddler Meddler = registry["identity"]
		for i, tag := range tags {
			tagLower := strings.ToLower(tag)
			//primary_key
			if strings.Contains(tagLower, "primary key") || strings.Contains(tagLower, "pk") {
				//主键不可以是指针
				if f.Type.Kind() == reflect.Ptr {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key but is a pointer", f.Name)
				}
				//primary key can only be one
				if data.pk != "" {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key, but a primary key field was already found", f.Name)
				}
				data.pk = name
			} else if m, present := registry[tag[j]]; present {
				// meddler = m
			} else {
				return nil, fmt.Errorf("meddler found field %s with meddler %s, but that meddler is not registered", f.Name, tag[j])
			}
		}
		if _, present := data.fields[name]; present {
			return nil, fmt.Errorf("meddler found multiple fields for column %s", name)
		}
		data.fields[name] = &structField{
			column:     name,
			primaryKey: name == data.pk,
			index:      i,
			meddler:    meddler,
		}
		data.columns = append(data.columns, name)
	}

	fieldsCache[dstType] = data
	return data, nil
}

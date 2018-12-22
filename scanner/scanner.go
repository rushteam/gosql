package scanner

import (
	"database/sql"
	"fmt"
	"log"
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

type StructField struct {
	column     string
	index      int
	primaryKey bool
	plugins    []string
}
type StructData struct {
	table   string
	columns []string
	fields  map[string]*StructField
	pk      string
	// ref     *reflect.Value
}

var Debug = false

//反射结构体缓存
var refStructCache = make(map[reflect.Type]*StructData)
var refStructCacheMutex sync.Mutex

//https://github.com/jmoiron/sqlx/blob/master/reflectx/reflect.go
func parseOptions(tag string) map[string]string {
	parts := strings.Split(tag, ",")
	options := make(map[string]string, len(parts))
	if len(parts) > 1 {
		for _, opt := range parts[1:] {
			// short circuit potentially expensive split op
			if strings.Contains(opt, "=") {
				kv := strings.Split(opt, "=")
				options[kv[0]] = kv[1]
				continue
			}
			options[opt] = ""
		}
	}
	return options
}
func ResolveStruct(dstType reflect.Type) (*StructData, error) {
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

	data := new(StructData)
	data.fields = make(map[string]*StructField)

	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		tags := strings.Split(f.Tag.Get(tagKey), tagSplit)
		// skip non-exported fields
		if f.PkgPath != "" && len(tags) < 1 {
			continue
		}
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
		field := &StructField{
			column:     name,
			primaryKey: name == data.pk,
			index:      i,
			// meddler:    meddler,
		}
		for _, tag := range tags {
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
			default:
				field.plugins = append(field.plugins, tag)
			}
		}
		if _, ok := data.fields[name]; ok {
			return nil, fmt.Errorf("scanner found multiple fields for column %s", name)
		}
		field.primaryKey = name == data.pk
		data.fields[name] = field
		data.columns = append(data.columns, name)
	}
	refStructCache[dstType] = data
	return data, nil
}
func scanRow(rows *sql.Rows, dst interface{}) error {
	if rows == nil {
		return fmt.Errorf("rows is a pointer, but not be a nil")
	}
	// get the sql columns
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	// check if there is data waiting
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	// maping struct address to  map address
	targets, err := Targets(dst, columns)
	if err != nil {
		return err
	}
	// perform the scan
	if err := rows.Scan(targets...); err != nil {
		return err
	}
	if len(columns) != len(targets) {
		return fmt.Errorf("scanner mismatch in number of columns (%d) and targets (%d)",
			len(columns), len(targets))
	}
	// format some field which have tag plugin
	if err := Plugins(dst, columns, targets); err != nil {
		return err
	}

	return rows.Err()
}
func Targets(dst interface{}, columns []string) ([]interface{}, error) {
	data, err := ResolveStruct(reflect.TypeOf(dst))
	if err != nil {
		return nil, err
	}
	structVal := reflect.ValueOf(dst).Elem()
	var targets []interface{}
	for _, name := range columns {
		if field, ok := data.fields[name]; ok {
			fieldAddr := structVal.Field(field.index).Addr().Interface()
			// fmt.Println(structVal.Field(field.index).Addr().Type())
			// scanTarget, err := field.meddler.PreRead(fieldAddr)
			if err != nil {
				return nil, fmt.Errorf("scanner.Targets: PreRead error on column %s: %v", name, err)
			}
			targets = append(targets, fieldAddr)
			// targets = append(targets, scanTarget)
		} else {
			// no destination, so throw this away
			targets = append(targets, new(interface{}))
			if Debug {
				log.Printf("scanner.Targets: column [%s] not found in struct", name)
			}
		}
	}
	return targets, nil
}

//https://github.com/russross/meddler/blob/038a8ef02b66198d4db78da3e9830fde52a7e072/meddler.go
func Plugins(dst interface{}, columns []string, targets []interface{}) error {
	data, err := ResolveStruct(reflect.TypeOf(dst))
	if err != nil {
		return err
	}
	structVal := reflect.ValueOf(dst).Elem()
	// for i, name := range data.columns {
	for i, name := range columns {
		if field, ok := data.fields[name]; ok {
			// field.Value.Addr().Interface()
			fieldAddr := structVal.Field(field.index).Addr().Interface()
			fmt.Println(i, name, fieldAddr, targets[i])
			_, _, _ = i, name, fieldAddr
			if err != nil {
				return fmt.Errorf("scanner.Plugins: PostRead error on column [%s]: %v", name, err)
			}
		} else {
			if Debug {
				log.Printf("scanner.Plugins: column [%s] not found in struct", name)
			}
		}
	}
	return nil
}
func Scan(rows *sql.Rows, dst interface{}) error {
	return scanRow(rows, dst)
}
func ScanAll(rows *sql.Rows, dst interface{}) error {
	defer rows.Close()
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("ScanAll called with non-pointer destination: %T", dst)
	}
	sliceVal := dstVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("ScanAll called with pointer to non-slice: %T", dst)
	}
	ptrType := sliceVal.Type().Elem()

	var eltType reflect.Type
	if ptrType.Kind() != reflect.Ptr {
		eltType = ptrType
	} else {
		eltType = ptrType.Elem()
	}
	if eltType.Kind() != reflect.Struct {
		return fmt.Errorf("ScanAll expects element to be pointers to structs, found %T", dst)
	}
	// gather the results
	for {
		// create a new element
		eltVal := reflect.New(eltType)
		elt := eltVal.Interface()
		// scan it
		if err := scanRow(rows, elt); err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}
		// add to the result slice
		if ptrType.Kind() != reflect.Ptr {
			sliceVal.Set(reflect.Append(sliceVal, eltVal.Elem()))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, eltVal))
		}
	}
}
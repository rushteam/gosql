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
const tagOptSplit = ":"
const tagColumn = "COLUMN"

//StructFieldOpts 模型字段选项
type StructFieldOpts map[string]string

//StructField 模型字段
type StructField struct {
	column       string
	index        int
	isPrimaryKey bool
	options      StructFieldOpts
	plugins      []string
}

//StructData 模型
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

//解析field tags to options
func parseTagOpts(tags reflect.StructTag) map[string]string {
	opts := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get(tagKey)} {
		if str != "" {
			tag := strings.Split(str, tagSplit)
			for i, value := range tag {
				v := strings.Split(value, tagOptSplit)
				k := strings.TrimSpace(strings.ToUpper(v[0]))
				if len(v) >= 2 {
					opts[k] = strings.Join(v[1:], tagOptSplit)
				} else {
					if i == 0 {
						opts[tagColumn] = v[0]
					} else {
						opts[k] = ""
					}
				}
			}
		}
	}
	return opts
}
func parseTableName(structType reflect.Type) string {
	return structType.Name()
}

//ResolveModelStruct 解析模型
func ResolveModelStruct(dstType reflect.Type) (*StructData, error) {
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
	data.table = parseTableName(structType)
	data.fields = make(map[string]*StructField)

	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		// skip non-exported fields
		if f.PkgPath != "" {
			continue
		}
		opts := parseTagOpts(f.Tag)
		// default to the field name
		column, ok := opts[tagColumn]
		if !ok {
			//todo 大小写转换下划线的、自定义方法的
			column = f.Name
		}
		// skip using "-" tag fields
		if column == "-" {
			continue
		}

		if _, ok := data.fields[column]; ok {
			return nil, fmt.Errorf("scanner found multiple fields for column %s", column)
		}
		for k, opt := range opts {
			switch k {
			case tagColumn:
				continue
			case "PK", "PRIMARY", "PRIMARY KEY", "PRIMARY_KEY":
				//pk can not is a pointer
				if f.Type.Kind() == reflect.Ptr {
					// indirectType := f.Type
					//indirectType = indirectType.Elem()
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key but is a pointer", f.Name)
				}
				//primary key can only be one
				if data.pk != "" {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key, but a primary key field was already found", f.Name)
				}
				data.pk = column
			case "UNI", "UNIQUE", "UNIQUE_INDEX":
				if opt == "" {
					opt = column
				}
				// data.uniques = append(data.uniques, opt)
			case "IDX", "INDEX":
				if opt == "" {
					opt = column
				}
			// data.indexs = append(data.indexs, opt)
			default:
				// if m, ok := marshalers[k]; ok {
				// field.marshaler =
				// }
			}
		}
		data.columns = append(data.columns, column)
		data.fields[column] = &StructField{
			column:       column,
			isPrimaryKey: column == data.pk,
			index:        i,
			options:      opts,
			// plugins:    plugins,
		}
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
	data, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		return nil, err
	}
	structVal := reflect.ValueOf(dst).Elem()
	//InterfaceSlice see http://code.google.com/p/go-wiki/wiki/InterfaceSlice
	// var targets []interface{}
	var targets = make([]interface{}, len(columns))
	for _, name := range columns {
		if field, ok := data.fields[name]; ok {
			//fieldAddr
			fieldValue := structVal.Field(field.index).Addr().Interface()
			// fmt.Println(structVal.Field(field.index).Addr().Type())
			// scanTarget, err := field.meddler.PreRead(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("scanner.Targets: PreRead error on column %s: %v", name, err)
			}
			switch fieldValue.(type) {
			// case sql.Scanner:
			//如果字段有scan方法 则调用
			// case *time.Time:
			// 	var scanAddr interface{}
			// 	scanAddr = new([]uint8)
			// 	targets = append(targets, scanAddr)
			default:
				targets = append(targets, fieldValue)
			}
			// targets = append(targets, fieldValue)
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
	data, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		return err
	}
	structVal := reflect.ValueOf(dst).Elem()
	// for i, name := range data.columns {
	for i, name := range columns {
		if field, ok := data.fields[name]; ok {
			// field.Value.Addr().Interface()
			fieldAddr := structVal.Field(field.index).Addr().Interface()
			// fmt.Println(targets[i])
			// fmt.Println(i, name, fieldAddr, targets[i])
			_, _, _ = i, name, fieldAddr
			// switch fieldAddr.(type) {
			// case *time.Time:
			// 	// fmt.Println(time.Unix(0, 0))
			// 	// fieldAddr = targets[i].(time.Time)
			// 	reflect.ValueOf(targets[i])
			// 	// fieldAddr = time.Unix(12312314, 0)
			// default:
			// 	// targets = append(targets, fieldAddr)
			// }
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

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
const tableFuncName = "TableName"

//Debug 模式
var Debug = false

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

//TableName ..
func (s StructData) TableName() string {
	if s.table == "" {
		panic("This method should be called before calling the ResolveModelTableName()")
	}
	return s.table
}

//Columns ..
func (s StructData) Columns() []string {
	return s.columns
}

//GetPk ..
func (s StructData) GetPk() string {
	return s.pk
}

//GetStructField ..
func (s StructData) GetStructField(k string) *StructField {
	if field, ok := s.fields[k]; ok {
		return field
	}
	return nil
}

//反射结构体缓存
var refStructCache = make(map[reflect.Type]*StructData)
var refStructCacheMutex sync.Mutex

//解析field tags to options
func parseTagOpts(tags reflect.StructTag) map[string]string {
	opts := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get(tagKey)} {
		if str == "" {
			continue
		}
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
	return opts
}

//UpdateModel ..
func UpdateModel(dst interface{}, list map[string]interface{}) {
	modelStruct, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		if Debug {
			log.Printf(err.Error())
		}
	}
	structVal := reflect.ValueOf(dst).Elem()
	for k, v := range list {
		if field, ok := modelStruct.fields[k]; ok {
			fmt.Println(structVal.Field(field.index).Kind())
			//reflect.Indirect(structVal.Field(field.index).Kind())
			fmt.Println(reflect.Indirect(structVal.Field(field.index)).Kind())

			if reflect.Indirect(structVal.Field(field.index)).Kind() != reflect.Indirect(reflect.ValueOf(v)).Kind() {
				log.Printf("[scanner.UpdateModel] value of type %s is not assignable to type %s",
					reflect.Indirect(reflect.ValueOf(v)).Kind(), reflect.Indirect(structVal.Field(field.index)).Kind())
				continue
			}

			//---------

			// switch reflect.Indirect(structVal.Field(field.index)).Kind() {
			// case reflect.String:
			// 	reflect.Indirect(structVal.Field(field.index)).Set(reflect.Indirect(reflect.ValueOf(v)))
			// case reflect.Bool:
			// 	reflect.Indirect(reflect.ValueOf(v)).Bool()
			// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// 	return value.Int() == 0
			// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			// 	return value.Uint() == 0
			// case reflect.Float32, reflect.Float64:
			// 	return value.Float() == 0
			// case reflect.Interface, reflect.Ptr:
			// 	return value.IsNil()
			// }
			reflect.Indirect(structVal.Field(field.index)).Set(reflect.Indirect(reflect.ValueOf(v)))
			// if structVal.Field(field.index).Kind() == reflect.Ptr {
			// 	structVal.Field(field.index).Elem().Set(reflect.ValueOf(v))
			// } else {
			// 	structVal.Field(field.index).Set(reflect.Indirect(reflect.ValueOf(v)))
			// }
		} else {
			if Debug {
				log.Printf("field [%s] not found in struct", k)
			}
		}
	}
	// listValue := reflect.ValueOf(list)
	// for _, field := range modelStruct.fields {
	// 	// if !structVal.Field(field.index).Addr().CanSet() {
	// 	// 	return fmt.Errorf("struct addr ")
	// 	// }
	// 	// fmt.Println(structVal.Field(field.index).Addr().CanSet())
	// 	// fmt.Println(structVal.Field(field.index).Addr().Elem().CanSet())
	// 	fieldValue := listValue.MapIndex(reflect.ValueOf(field.column))
	// 	if fieldValue.IsValid() {
	// 		structVal.Field(field.index).Addr().Elem().Set(listValue.MapIndex(reflect.ValueOf(field.column)))
	// 	}
	// }
}

//ResolveModelToMap 解析模型数据到 非零值不解析
func ResolveModelToMap(dst interface{}) (map[string]interface{}, error) {
	var list = make(map[string]interface{}, 0)
	modelStruct, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		return list, err
	}
	structVal := reflect.ValueOf(dst).Elem()
	for _, field := range modelStruct.fields {
		if structVal.Field(field.index).Kind() == reflect.Ptr {
			if structVal.Field(field.index).IsNil() {
				list[field.column] = reflect.New(structVal.Field(field.index).Type()).Interface()
			} else {
				list[field.column] = structVal.Field(field.index).Elem().Interface()
			}
		} else if !isZeroVal(structVal.Field(field.index)) {
			list[field.column] = structVal.Field(field.index).Addr().Interface()
		}
	}
	return list, nil
}

//ResolveModelTableName 解析方法名
func ResolveModelTableName(dst interface{}) (string, error) {
	dstType, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		return "", err
	}
	if dstType.table == "" {
		structVal := reflect.ValueOf(dst).Elem()
		name := structVal.Type().Name()
		fnTableName := structVal.MethodByName(tableFuncName)
		if fnTableName.IsValid() {
			name = fnTableName.Call([]reflect.Value{})[0].Interface().(string)
		}
		dstType.table = name
		//todo 大小写转换下划线的、自定义方法的
		// if TableNameFormat == TableNameSnake {
		// 	name = utils.SnakeString(name)
		// }
	}
	return dstType.table, nil
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
	//这里不方便获取到方法自定义method上的table名所以滞后到ResolveModelTableName中
	// data.table = dstType.Elem().Name()
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
		if !ok || column == "" {
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

//Targets ..
func Targets(dst interface{}, columns []string) ([]interface{}, error) {
	data, err := ResolveModelStruct(reflect.TypeOf(dst))
	if err != nil {
		return nil, err
	}
	structVal := reflect.ValueOf(dst).Elem()
	//InterfaceSlice see http://code.google.com/p/go-wiki/wiki/InterfaceSlice
	//var targets = []interface{} targets = append(targets, fieldValue)
	var targets = make([]interface{}, len(columns))
	for i, name := range columns {
		if field, ok := data.fields[name]; ok {
			//fieldAddr
			fieldValue := structVal.Field(field.index).Addr().Interface()
			// fmt.Println(structVal.Field(field.index).Addr().Type())
			// scanTarget, err := field.meddler.PreRead(fieldValue)
			// if err != nil {
			// 	return nil, fmt.Errorf("scanner.Targets: PreRead error on column %s: %v", name, err)
			// }
			switch fieldValue.(type) {
			// case sql.Scanner:
			//如果字段有scan方法 则调用
			// case *time.Time:
			// 	var scanAddr interface{}
			// 	scanAddr = new([]uint8)
			// 	targets = append(targets, scanAddr)
			default:
				targets[i] = fieldValue
			}
			// targets = append(targets, fieldValue)
			// targets = append(targets, fieldValue)
			// targets = append(targets, scanTarget)
		} else {
			// no destination, so throw this away
			// targets[i] = new(interface{})
			targets[i] = new(sql.RawBytes)
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
func isZeroVal(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

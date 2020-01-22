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
		panic("scanner not foundd table name")
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
var (
	refStructCache      = make(map[reflect.Type]*StructData)
	refStructCacheMutex sync.Mutex
)

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
	modelStruct, err := ResolveModelStruct(dst)
	if err != nil {
		if Debug {
			log.Printf(err.Error())
		}
	}
	structRV := reflect.ValueOf(dst).Elem()
	// fmt.Printf("%v", structVal.CanSet())
	for k, v := range list {
		if field, ok := modelStruct.fields[k]; ok {
			if structRV.Field(field.index).Kind() == reflect.Ptr {
				if structRV.Field(field.index).IsNil() && structRV.Field(field.index).CanSet() {
					structRV.Field(field.index).Set(reflect.New(structRV.Field(field.index).Type().Elem()))
				}
				structRV.Field(field.index).Elem().Set(reflect.ValueOf(v))
			} else {
				structRV.Field(field.index).Set(reflect.Indirect(reflect.ValueOf(v)))
			}
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
	structRV := reflect.ValueOf(dst).Elem()
	//如果是数组则忽略
	if structRV.Kind() == reflect.Slice {
		return list, nil
	}
	modelStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return list, err
	}
	for _, field := range modelStruct.fields {
		if structRV.Field(field.index).Kind() == reflect.Ptr {
			//忽略掉指针为nil 和 零值情况
			if structRV.Field(field.index).IsNil() {
				//指针为nil时候的处理
				continue
				// list[field.column] = reflect.New(structRV.Field(field.index).Type()).Interface()
				// list[field.column] = sql.NullString
			}
			if structRV.Field(field.index).Elem().IsZero() || structRV.Field(field.index).Elem().IsValid() {
				continue
			}
			list[field.column] = structRV.Field(field.index).Elem().Interface()
			continue
		}
		// if isZeroVal(structRV.Field(field.index)) {
		// 	continue
		// }
		if structRV.Field(field.index).IsZero() {
			continue
		}
		fmt.Println("++", field.column, structRV.Field(field.index))
		list[field.column] = structRV.Field(field.index).Addr().Interface()
	}
	return list, nil
}

//SnakeString  转 snake_string
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

//FormatName 格式化字段名
func FormatName(name string) string {
	return SnakeString(name)
}

//ResolveModelStruct 解析模型
func ResolveModelStruct(dst interface{}) (*StructData, error) {
	var structRV reflect.Value
	var structRT reflect.Type
	modelStruct := new(StructData)
	dstRV := reflect.ValueOf(dst)
	//兼容指针逻辑
	if dstRV.Kind() == reflect.Ptr {
		dstRV = dstRV.Elem()
	}
	//dst (slice)
	if dstRV.Kind() == reflect.Slice {
		ptrRT := dstRV.Type().Elem()
		if ptrRT.Kind() == reflect.Ptr {
			ptrRT = ptrRT.Elem()
		}
		if ptrRT.Kind() != reflect.Struct {
			return nil, fmt.Errorf("scanner expects element to be pointers to structs, found %T", dst)
		}
		structRT = ptrRT
		structRV = reflect.New(structRT)
	} else {
		//dst (struct)
		structRT = reflect.TypeOf(dst)
		refStructCacheMutex.Lock()
		defer refStructCacheMutex.Unlock()

		if modelStruct, ok := refStructCache[structRT]; ok {
			return modelStruct, nil
		}
		if structRT.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("scanner called with non-pointer destination %v", structRT)
		}
		structRT = structRT.Elem() //struct
		if structRT.Kind() != reflect.Struct {
			return nil, fmt.Errorf("scanner called with pointer to non-struct %v", structRT)
		}

		// modelStruct.table = structRT.Elem().Name()
		//兼容
		if reflect.ValueOf(dst).IsValid() {
			structRV = reflect.New(structRT)
		} else {
			structRV = reflect.ValueOf(dst).Elem()
		}
	}
	//这里从value上获取到自定义method上的table name
	fnTableName := structRV.MethodByName(tableFuncName)
	if fnTableName.IsValid() {
		modelStruct.table = fnTableName.Call([]reflect.Value{})[0].Interface().(string)
	} else {
		modelStruct.table = structRV.Type().Name()
		//todo 大小写转换下划线的、自定义方法的
		// if TableNameFormat == TableNameSnake {
		// 	name = utils.SnakeString(name)
		// }
	}
	modelStruct.fields = make(map[string]*StructField)

	for i := 0; i < structRT.NumField(); i++ {
		f := structRT.Field(i)
		// skip non-exported fields
		if f.PkgPath != "" {
			continue
		}
		opts := parseTagOpts(f.Tag)
		// default to the field name
		column, ok := opts[tagColumn]
		if !ok || column == "" {
			//todo 大小写转换下划线的、自定义方法的
			column = FormatName(f.Name)
		}
		// skip using "-" tag fields
		if column == "-" {
			continue
		}

		if _, ok := modelStruct.fields[column]; ok {
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
				if modelStruct.pk != "" {
					return nil, fmt.Errorf("scanner found field %s which is marked as the primary key, but a primary key field was already found", f.Name)
				}
				modelStruct.pk = column
			case "UNI", "UNIQUE", "UNIQUE_INDEX":
				if opt == "" {
					opt = column
				}
				// modelStruct.uniques = append(modelStruct.uniques, opt)
			case "IDX", "INDEX":
				if opt == "" {
					opt = column
				}
			// modelStruct.indexs = append(modelStruct.indexs, opt)
			default:
				// if m, ok := marshalers[k]; ok {
				// field.marshaler =
				// }
			}
		}
		modelStruct.columns = append(modelStruct.columns, column)
		modelStruct.fields[column] = &StructField{
			column:       column,
			isPrimaryKey: column == modelStruct.pk,
			index:        i,
			options:      opts,
			// plugins:    plugins,
		}
	}
	refStructCache[structRT] = modelStruct
	return modelStruct, nil
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
	dstStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
	// dstRT := reflect.TypeOf(dst)
	// dstRT.Field()
	dstRV := reflect.ValueOf(dst)
	//兼容指针逻辑
	if dstRV.Kind() == reflect.Ptr {
		dstRV = dstRV.Elem()
	}
	//InterfaceSlice see http://code.google.com/p/go-wiki/wiki/InterfaceSlice
	var targets = make([]interface{}, len(columns))
	for i, name := range columns {
		if field, ok := dstStruct.fields[name]; ok {
			if dstRV.Field(field.index).CanAddr() {
				fieldValue := dstRV.Field(field.index).Addr().Interface()
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
			} else {
				targets[i] = new(interface{})
				// targets[i] = new(sql.RawBytes)
			}
		} else {
			//结构体不存在这个字段时候
			// targets[i] = new(sql.RawBytes)
			targets[i] = new(interface{})
			if Debug {
				log.Printf("scanner.Targets: column [%s] not found in struct", name)
			}
		}
	}
	return targets, nil
}

//https://github.com/russross/meddler/blob/038a8ef02b66198d4db78da3e9830fde52a7e072/meddler.go
func Plugins(dst interface{}, columns []string, targets []interface{}) error {
	dstStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return err
	}
	dstRV := reflect.ValueOf(dst)
	//兼容指针逻辑
	if dstRV.Kind() == reflect.Ptr {
		dstRV = dstRV.Elem()
	}
	// for i, name := range data.columns {
	for i, name := range columns {
		if field, ok := dstStruct.fields[name]; ok {
			if dstRV.Field(field.index).CanAddr() {
				fieldAddr := dstRV.Field(field.index).Addr().Interface()
				_, _, _ = i, name, fieldAddr
			}
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

//Scan ..
func Scan(rows *sql.Rows, dst interface{}) error {
	return scanRow(rows, dst)
}

//ScanAll ..
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

//使用内在方法
// func isZeroVal(value reflect.Value) bool {
// 	switch value.Kind() {
// 	case reflect.String:
// 		return value.Len() == 0
// 	case reflect.Bool:
// 		return !value.Bool()
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return value.Int() == 0
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
// 		return value.Uint() == 0
// 	case reflect.Float32, reflect.Float64:
// 		return value.Float() == 0
// 	case reflect.Interface, reflect.Ptr:
// 		return value.IsNil()
// 	}
// 	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
// }

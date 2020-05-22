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

//TagKey parse struct tag
const TagKey = "db"
const tagSplit = ","
const tagOptSplit = ":"
const tagColumn = "COLUMN"
const tableFuncName = "TableName"

//Debug 模式
var Debug = false

//StructField 模型字段
type StructField struct {
	column       string
	index        int
	isPrimaryKey bool
	//模型字段选项
	options map[string]string
	plugins []string
}

//StructData 模型
type StructData struct {
	table   string
	columns []string
	fields  map[string]*StructField
	pk      string
	Uniques []string
	Indexs  []string
	// ref     *reflect.Value
}

//TableName ..
func (s StructData) TableName() string {
	if s.table == "" {
		panic("table name can not be empty")
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
	for _, str := range []string{tags.Get("sql"), tags.Get(TagKey)} {
		if str == "" {
			continue
		}
		tag := strings.Split(str, tagSplit)
		for i, value := range tag {
			kv := strings.Split(value, tagOptSplit)
			k := strings.TrimSpace(strings.ToUpper(kv[0]))
			if len(kv) >= 2 { //eg: `db:"column:id"`
				opts[k] = strings.Join(kv[1:], tagOptSplit)
			} else {
				if i == 0 { //`db:"id"`
					opts[tagColumn] = value
				} else { //`db:"id,xx"`
					opts[k] = ""
				}
			}
		}
	}
	return opts
}

//UpdateModel ..
func UpdateModel(dst interface{}, list map[string]interface{}) error {
	modelStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return err
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
	return nil
}

func getStructVal(structRV reflect.Value, index int) interface{} {
	if structRV.Field(index).Kind() == reflect.Ptr {
		//忽略掉指针为nil 和 零值情况
		if structRV.Field(index).IsNil() {
			//指针为nil时候的处理
			return nil
			//return reflect.New(structRV.Field(field.index).Type()).Interface()
		}
		if structRV.Field(index).Elem().IsZero() || structRV.Field(index).Elem().IsValid() {
			return nil
		}
		return structRV.Field(index).Elem().Interface()
	}
	if structRV.Field(index).IsZero() {
		return nil
	}
	return structRV.Field(index).Interface()
}

//ResolveStructValue 解析模型数据到 非零值不解析
func ResolveStructValue(dst interface{}) (map[string]interface{}, error) {
	var list = make(map[string]interface{}, 0)
	dstRV := reflect.ValueOf(dst)
	//兼容指针逻辑
	if dstRV.Kind() == reflect.Ptr {
		dstRV = dstRV.Elem()
	}
	//如果是数组则忽略
	// if dstRV.Kind() == reflect.Slice {
	// 	return nil, nil
	// }
	if dstRV.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Must be a struct, scanner called with non-struct: %v", dstRV.Kind())
	}
	modelStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return list, err
	}
	for _, field := range modelStruct.fields {
		v := getStructVal(dstRV, field.index)
		if v != nil {
			list[field.column] = v
		}
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

//resolveStruct struct resolve to StructData
func resolveStruct(structRV reflect.Value) (*StructData, error) {
	var structRT reflect.Type
	structRT = structRV.Type()
	if modelStruct, ok := refStructCache[structRT]; ok {
		return modelStruct, nil
	}
	modelStruct := new(StructData)
	//这里从value上获取到自定义method上的table name
	var fnTableName reflect.Value
	if structRV.CanAddr() {
		fnTableName = structRV.Addr().MethodByName(tableFuncName)
	} else {
		fnTableName = structRV.MethodByName(tableFuncName)
	}
	if fnTableName.IsValid() {
		modelStruct.table = fnTableName.Call([]reflect.Value{})[0].Interface().(string)
	} else {
		//todo 大小写转换下划线的、自定义方法的
		// if TableNameFormat == TableNameSnake {
		// 	name = utils.SnakeString(name)
		// }
		modelStruct.table = structRT.Name()
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
				modelStruct.Uniques = append(modelStruct.Uniques, opt)
			case "IDX", "INDEX":
				if opt == "" {
					opt = column
				}
				modelStruct.Indexs = append(modelStruct.Indexs, opt)
			default:
				// if m, ok := marshalers[k]; ok {
				// field.marshaler =
				// }
			}
		}
		//未指定情况下寻找id名称的字段
		if modelStruct.pk == "" && column == "id" {
			modelStruct.pk = column
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
	refStructCacheMutex.Lock()
	defer refStructCacheMutex.Unlock()
	refStructCache[structRT] = modelStruct
	return modelStruct, nil
}

//resolveModel model resolve to StructData
func resolveModel(dstRV reflect.Value) (*StructData, error) {
	switch dstRV.Kind() {
	case reflect.Struct:
		return resolveStruct(dstRV)
	case reflect.Ptr:
		if dstRV.IsZero() {
			dstRV = reflect.New(dstRV.Type().Elem())
		}
		dstRV = dstRV.Elem()
		return resolveModel(dstRV)
	case reflect.Slice:
		eltRT := dstRV.Type().Elem()
		eltRV := reflect.New(eltRT)
		return resolveModel(eltRV)
	default:
		return nil, fmt.Errorf("scanner expects pointer must pointers to struct or slice, found %v", dstRV.Kind())
	}
}

//ResolveModelStruct 解析模型
func ResolveModelStruct(dst interface{}) (*StructData, error) {
	dstRV := reflect.ValueOf(dst)
	switch dstRV.Kind() {
	case reflect.Ptr:
		dstRV = dstRV.Elem()
		return resolveModel(dstRV)
	case reflect.Slice, reflect.Struct:
		return resolveModel(dstRV)
	}
	return nil, fmt.Errorf("scanner expects pointer must pointers to struct or slice, found %v", dstRV.Kind())
}

//Targets ..
func Targets(dst interface{}, columns []string) ([]interface{}, error) {
	dstStruct, err := ResolveModelStruct(dst)
	if err != nil {
		return nil, err
	}
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

//Plugins ..
//see https://github.com/russross/meddler/blob/038a8ef02b66198d4db78da3e9830fde52a7e072/meddler.go
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
	if rows == nil {
		return fmt.Errorf("rows is a pointer, but not be a nil")
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	// bind struct-address to map-address
	targets, err := Targets(dst, columns)
	if err != nil {
		return err
	}
	if len(columns) != len(targets) {
		return fmt.Errorf("scanner mismatch in number of columns (%d) and targets (%d)",
			len(columns), len(targets))
	}
	if err := rows.Scan(targets...); err != nil {
		return err
	}
	// format some field which have tag plugin
	if err := Plugins(dst, columns, targets); err != nil {
		return err
	}
	return rows.Err()
}

//ScanRow ScanRow and Close Rows
func ScanRow(rows *sql.Rows, dst interface{}) error {
	defer rows.Close()
	return Scan(rows, dst)
}

//ScanAll ..
func ScanAll(rows *sql.Rows, dst interface{}) error {
	defer rows.Close()
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("ScanAll called with non-pointer destination: %T", dst)
	}
	sliceRV := dstVal.Elem()
	if sliceRV.Kind() != reflect.Slice {
		return fmt.Errorf("ScanAll called with pointer to non-slice: %T", dst)
	}
	sliceRT := sliceRV.Type()
	eltRT := sliceRT.Elem()
	if eltRT.Kind() == reflect.Ptr {
		eltRT = eltRT.Elem()
	}
	if eltRT.Kind() != reflect.Struct {
		return fmt.Errorf("ScanAll expects element to be pointers to structs, found %T", dst)
	}
	// gather the results
	for {
		// create a new element
		eltRV := reflect.New(eltRT)
		elt := eltRV.Interface()
		// scan it
		if err := Scan(rows, elt); err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}
		// add struct to slice
		if sliceRT.Elem().Kind() == reflect.Ptr {
			sliceRV.Set(reflect.Append(sliceRV, eltRV))
		} else {
			sliceRV.Set(reflect.Append(sliceRV, eltRV.Elem()))
		}
	}
}

package gosql

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const (
	_select uint8 = iota
	_insert
	_replace
	_update
	_delete
)

var tagKey = "db"
var identKey = "`"

//SQLSegments ...
type SQLSegments struct {
	table   []TbName
	fields  []string
	flags   []string
	join    []map[string]string
	where   Clause
	groupBy []string
	having  Clause
	orderBy []string
	limit   struct {
		limit  int
		offset int
	}
	union     []func(*SQLSegments)
	forUpdate bool
	returning bool
	// params    []interface{}
	params []map[string]interface{}
	render struct {
		args []interface{}
	}
	//sql cmd type: select|insert|repalce|update|delete
	cmd uint8
}

//TbName ..
type TbName struct {
	Name  string
	Alias string
}

//Add ..
type Add int

//Sub ..
type Sub int

//NewSQLSegment ..
func NewSQLSegment() *SQLSegments {
	return &SQLSegments{}
}

//Table SQLSegments
func (s *SQLSegments) Table(name interface{}) *SQLSegments {
	switch v := name.(type) {
	case TbName:
		s.table = append(s.table, v)
	case []TbName:
		s.table = append(s.table, v...)
	case string:
		s.table = append(s.table, TbName{v, ""})
	}
	return s
}

//Field SQLSegments
func (s *SQLSegments) Field(fields ...string) *SQLSegments {
	if len(fields) > 0 {
		s.fields = append(s.fields, fields...)
	}
	return s
}

//Flag SQLSegments
func (s *SQLSegments) Flag(flags ...string) *SQLSegments {
	if len(flags) > 0 {
		s.flags = append(s.flags, flags...)
	}
	return s
}

//Join SQLSegments
func (s *SQLSegments) Join(table string, conditionA, logic, conditionB string) *SQLSegments {
	s.addJoin("", table, conditionA, logic, conditionB)
	return s
}

//LeftJoin SQLSegments
func (s *SQLSegments) LeftJoin(table string, conditionA, logic, conditionB string) *SQLSegments {
	s.addJoin("LEFT", table, conditionA, logic, conditionB)
	return s
}

//RightJoin SQLSegments
func (s *SQLSegments) RightJoin(table string, conditionA, logic, conditionB string) *SQLSegments {
	s.addJoin("RIGHT", table, conditionA, logic, conditionB)
	return s
}

//InnerJoin SQLSegments
func (s *SQLSegments) InnerJoin(table string, conditionA, logic, conditionB string) *SQLSegments {
	s.addJoin("INNER", table, conditionA, logic, conditionB)
	return s
}

//CorssJoin SQLSegments
func (s *SQLSegments) CorssJoin(table string, conditionA, logic, conditionB string) *SQLSegments {
	s.addJoin("CROSS", table, conditionA, logic, conditionB)
	return s
}

//addJoin SQLSegments
func (s *SQLSegments) addJoin(typ string, table string, conditionA, logic, conditionB string) *SQLSegments {
	var t = make(map[string]string)
	t["type"] = typ
	t["table"] = table
	t["logic"] = logic
	t["conditionA"] = conditionA
	t["conditionB"] = conditionB
	s.join = append(s.join, t)
	return s
}

//OrderBy SQLSegments
func (s *SQLSegments) OrderBy(fields ...string) *SQLSegments {
	if len(fields) > 0 {
		s.orderBy = append(s.orderBy, fields...)
	}
	return s
}

//GroupBy SQLSegments
func (s *SQLSegments) GroupBy(fields ...string) *SQLSegments {
	if len(fields) > 0 {
		s.groupBy = append(s.groupBy, fields...)
	}
	return s
}

//Offset SQLSegments
func (s *SQLSegments) Offset(n int) *SQLSegments {
	s.limit.offset = n
	return s
}

//Limit SQLSegments
func (s *SQLSegments) Limit(n int) *SQLSegments {
	s.limit.limit = n
	return s
}

//ForUpdate SQLSegments
func (s *SQLSegments) ForUpdate() *SQLSegments {
	s.forUpdate = true
	return s
}

//Returning SQLSegments
func (s *SQLSegments) Returning() *SQLSegments {
	s.returning = true
	return s
}

//Clause ...
type Clause struct {
	key    interface{}
	val    interface{}
	logic  string
	clause []*Clause
}

func (p *Clause) addClause(logic string, key interface{}, vals ...interface{}) *Clause {
	var c = &Clause{}
	c.logic = logic
	switch k := key.(type) {
	case func(*Clause):
		k(c)
		// p.clause = append(p.clause, c)
	default:
		c.key = key
		if len(vals) > 0 {
			c.val = vals[0]
		}
	}
	// fmt.Println(p.clause)
	p.clause = append(p.clause, c)
	return p
}

//Where ..
func (p *Clause) Where(key interface{}, vals ...interface{}) *Clause {
	p.addClause("AND", key, vals...)
	return p
}

//OrWhere ..
func (p *Clause) OrWhere(key interface{}, vals ...interface{}) *Clause {
	p.addClause("OR", key, vals...)
	return p
}

//Build ...
func (p *Clause) Build(i int) (string, []interface{}) {
	var sql = ""
	var args []interface{}
	if p.logic != "" && i > 0 {
		sql += " " + p.logic
	}
	switch k := p.key.(type) {
	case string:
		r, _ := regexp.Compile(`\[(\>\=|\<\=|\>|\<|\<\>|\!\=|\=|\~|\!\~|like|!like|in|!in|is|!is|exists|!exists|#)\]?([a-zA-Z0-9_.\-\=\s\?\(\)]*)`)
		match := r.FindStringSubmatch(k)
		var context string
		if len(match) > 0 {
			// fmt.Println(len(match), match[1])
			switch match[1] {
			case "~", "like":
				context = buildIdent(match[2]) + " LIKE ?"
				args = append(args, p.val)
			case "!~", "!like":
				context = buildIdent(match[2]) + "` NOT LIKE ?"
				args = append(args, p.val)
			case ">":
				context = buildIdent(match[2]) + " > ?"
				args = append(args, p.val)
			case ">=":
				context = buildIdent(match[2]) + "` >= ?"
				args = append(args, p.val)
			case "<":
				context = buildIdent(match[2]) + " < ?"
				args = append(args, p.val)
			case "<=":
				context = buildIdent(match[2]) + " <= ?"
				args = append(args, p.val)
			case "<>", "!=":
				context = buildIdent(match[2]) + " != ?"
				args = append(args, p.val)
			case "=":
				context = buildIdent(match[2]) + " = ?"
				args = append(args, p.val)
			case "in":
				context = buildIdent(match[2]) + " IN ("
				var holder string
				if reflect.TypeOf(p.val).Kind() == reflect.Slice {
					v := reflect.ValueOf(p.val)
					holder = buildPlaceholder(v.Len(), "?", " ,")
					for n := 0; n < v.Len(); n++ {
						args = append(args, v.Index(n).Interface())
					}
				} else {
					holder = "?"
					args = append(args, p.val)
				}
				context += holder + ")"
			case "!in":
				context = buildIdent(match[2]) + " NOT IN ("
				var holder string
				if reflect.TypeOf(p.val).Kind() == reflect.Slice {
					v := reflect.ValueOf(p.val)
					holder = buildPlaceholder(v.Len(), "?", " ,")
					for n := 0; n < v.Len(); n++ {
						args = append(args, v.Index(n).Interface())
					}
				} else {
					holder = "?"
					args = append(args, p.val)
				}
				context += holder + ")"
			case "exists":
				switch p.val.(type) {
				case string:
					context = "EXISTS (" + p.val.(string) + ")"
				case func(s *SQLSegments):
					s := NewSQLSegment()
					p.val.(func(s *SQLSegments))(s)
					context = "EXISTS (" + s.BuildSelect() + ")"
					args = append(args, s.render.args...)
				}
			case "!exists":
				switch p.val.(type) {
				case string:
					context = "NOT EXISTS (" + p.val.(string) + ")"
				case func(s *SQLSegments):
					s := NewSQLSegment()
					p.val.(func(s *SQLSegments))(s)
					context = "NOT EXISTS (" + s.BuildSelect() + ")"
					args = append(args, s.render.args...)
				}
			case "is":
				if p.val == nil {
					context = buildIdent(match[2]) + " IS NULL"
				} else {
					context = buildIdent(match[2]) + " IS ?"
					args = append(args, p.val)
				}
			case "!is":
				if p.val == nil {
					context = buildIdent(match[2]) + " IS NOT NULL"
				} else {
					context = buildIdent(match[2]) + " IS NOT ?"
					args = append(args, p.val)
				}
			case "#":
				context = match[2]
				if reflect.TypeOf(p.val).Kind() == reflect.Slice {
					v := reflect.ValueOf(p.val)
					for n := 0; n < v.Len(); n++ {
						args = append(args, v.Index(n).Interface())
					}
				} else {
					args = append(args, p.val)
				}
			}
			sql += " " + context
		} else {
			if p.val != nil {
				sql += " " + buildIdent(k) + " = ?"
				args = append(args, p.val)
			} else {
				sql += " " + k
			}
		}
	case nil:
		sql += " ("
		for j, c := range p.clause {
			part, arg := c.Build(j)
			sql += part
			args = append(args, arg...)
		}
		sql += ")"
	}
	return sql, args
}

//Where ..
func (s *SQLSegments) Where(key interface{}, vals ...interface{}) *SQLSegments {
	s.where.Where(key, vals...)
	return s
}

//OrWhere ..
func (s *SQLSegments) OrWhere(key interface{}, vals ...interface{}) *SQLSegments {
	s.where.OrWhere(key, vals...)
	return s
}

//BuildWhereClause ...
func (s *SQLSegments) buildWhereClause() string {
	var sql string
	if len(s.where.clause) > 0 {
		sql = " WHERE"
		for i, c := range s.where.clause {
			part, args := c.Build(i)
			sql += part
			s.render.args = append(s.render.args, args...)
		}
	}
	return sql
}

//IsEmptyWhereClause ...
func (s *SQLSegments) IsEmptyWhereClause() bool {
	return len(s.where.clause) < 1
}

//Having ...
func (s *SQLSegments) Having(key interface{}, vals ...interface{}) *SQLSegments {
	s.having.Where(key, vals...)
	return s
}

//buildHavingClause ...
func (s *SQLSegments) buildHavingClause() string {
	var sql string
	if len(s.having.clause) > 0 {
		sql = " HAVING"
		for i, c := range s.having.clause {
			part, args := c.Build(i)
			sql += part
			s.render.args = append(s.render.args, args...)
		}
	}
	return sql
}
func (s *SQLSegments) buildFlags() string {
	var sql string
	for _, v := range s.flags {
		sql += " " + v
	}
	return sql

}
func (s *SQLSegments) buildField() string {
	var sql string
	if len(s.fields) == 0 {
		sql += " *"
	} else {
		for i, v := range s.fields {
			if i > 0 {
				sql += ","
			}
			if v == "*" {
				sql += " " + v
			} else {
				sql += " " + buildIdent(v)
			}
		}
	}
	return sql
}
func (s *SQLSegments) buildTable() string {
	var sql string
	for i, v := range s.table {
		if i > 0 {
			sql += ","
		}
		sql += " " + buildIdent(v.Name)
		if v.Alias != "" {
			sql += " AS " + buildIdent(v.Alias)
		}
	}
	return sql
}
func (s *SQLSegments) buildJoin() string {
	var sql string
	for _, t := range s.join {
		sql += " " + t["type"] + "JOIN " + buildIdent(t["table"]) + " ON " + buildIdent(t["conditionA"]) + " " + t["logic"] + " " + buildIdent(t["conditionB"])
	}
	return sql
}
func (s *SQLSegments) buildGroupBy() string {
	var sql string
	if len(s.groupBy) > 0 {
		sql += " GROUP BY"
	}
	for i, v := range s.groupBy {
		if i > 0 {
			sql += ","
		}
		sql += " " + buildIdent(v)
	}
	return sql
}
func (s *SQLSegments) buildOrderBy() string {
	var sql string
	if len(s.orderBy) > 0 {
		sql += " ORDER BY"
	}
	for i, v := range s.orderBy {
		if i > 0 {
			sql += ","
		}
		sql += " " + buildIdent(v)
	}
	return sql
}
func (s *SQLSegments) buildLimit() string {
	var sql string
	if s.limit.limit != 0 {
		sql += fmt.Sprintf(" LIMIT %d", s.limit.limit)
	}
	if s.limit.offset != 0 {
		sql += fmt.Sprintf(" OFFSET %d", s.limit.offset)
	}
	return sql
}

//Union ...
func (s *SQLSegments) Union(f func(*SQLSegments)) *SQLSegments {
	s.union = append(s.union, f)
	return s
}
func (s *SQLSegments) buildUnion() string {
	var sql string
	if len(s.union) > 0 {
		sql += " UNION ("
	}
	for _, f := range s.union {
		var ss = &SQLSegments{}
		f(ss)
		sql += ss.BuildSelect()
	}
	if len(s.union) > 0 {
		sql += ")"
	}
	return sql
}
func (s *SQLSegments) buildForUpdate() string {
	if s.forUpdate == true {
		return " FOR UPDATE"
	}
	return ""
}

//BuildSelect ...
func (s *SQLSegments) BuildSelect() string {
	var sql = fmt.Sprintf("SELECT%s%s FROM%s%s%s%s%s%s%s%s%s",
		s.buildFlags(),
		s.buildField(),
		s.buildTable(),
		s.buildJoin(),
		s.buildWhereClause(),
		s.buildGroupBy(),
		s.buildHavingClause(),
		s.buildOrderBy(),
		s.buildLimit(),
		s.buildUnion(),
		s.buildForUpdate(),
	)
	s.cmd = _select
	// fmt.Println(s.render.args)
	return sql
}

//Params ...
func (s *SQLSegments) Params(vals ...map[string]interface{}) *SQLSegments {
	s.params = append(s.params, vals...)
	return s
}

//Insert ...
func (s *SQLSegments) Insert(vals ...map[string]interface{}) *SQLSegments {
	s.params = append(s.params, vals...)
	return s
}

//BuildInsert ...
func (s *SQLSegments) BuildInsert() string {
	var sql = fmt.Sprintf("INSERT%s INTO%s%s%s",
		s.buildFlags(),
		s.buildTable(),
		s.buildValuesForInsert(),
		s.buildReturning(),
	)
	s.cmd = _insert
	return sql
}

//BuildReplace ...
func (s *SQLSegments) BuildReplace() string {
	var sql = fmt.Sprintf("REPLACE%s INTO%s%s%s",
		s.buildFlags(),
		s.buildTable(),
		s.buildValuesForInsert(),
		s.buildReturning(),
	)
	s.cmd = _replace
	return sql
}

//BuildInsert ...
func (s *SQLSegments) buildValuesForInsert() string {
	var fields string
	var values string
	var fieldSlice []string
	for i, vals := range s.params {
		if i == 0 {
			for arg := range vals {
				fieldSlice = append(fieldSlice, arg)
			}
		}
	}
	fieldLen := len(fieldSlice)
	fields += buildString(fieldSlice, ",", " (", ")", true)
	for i, vals := range s.params {
		if i == 0 {
			values += " ("
		} else {
			values += ",("
		}
		for _, arg := range fieldSlice {
			s.render.args = append(s.render.args, vals[arg])
		}
		values += buildPlaceholder(fieldLen, "?", ",")
		values += ")"
	}
	var sql = fields + " VALUES" + values
	return sql
}

//UpdateField 更新字段
func (s *SQLSegments) UpdateField(key string, val interface{}) *SQLSegments {
	if len(s.params) == 0 {
		s.params = append(s.params, make(map[string]interface{}, 0))
	}
	//update set 值只能是一行的维度（只用params[0]）
	s.params[0][key] = val
	return s
}

//Update ..
func (s *SQLSegments) Update(vals map[string]interface{}) *SQLSegments {
	//panic("Update method only one parameter is supported")
	if len(vals) < 1 {
		panic("Must be have values")
	}
	s.params = append(s.params, vals)
	return s
}

//UnsafeUpdate 可以没有where条件更新 ,Update 更新必须指定where条件才能更新否则panic
func (s *SQLSegments) UnsafeUpdate(vals map[string]interface{}) *SQLSegments {
	s.params = append(s.params, vals)
	return s
}

//buildReturning ...
func (s *SQLSegments) buildReturning() string {
	if s.returning == true {
		return " RETURNING"
	}
	return ""
}

//BuildUpdate ...
func (s *SQLSegments) BuildUpdate() string {
	var sql = fmt.Sprintf("UPDATE%s%s%s%s%s%s%s",
		s.buildFlags(),
		s.buildTable(),
		s.buildValuesForUpdate(),
		s.buildWhereClause(),
		s.buildOrderBy(),
		s.buildLimit(),
		s.buildReturning(),
	)
	s.cmd = _update
	// fmt.Println(s.render.args)
	return sql
}

//buildValuesForUpdate ...
func (s *SQLSegments) buildValuesForUpdate() string {
	var buffer bytes.Buffer
	buffer.WriteString(" SET ")
	// var fieldSlice []string
	if len(s.params) == 0 {
		panic(fmt.Sprintf("Must be have values after 'UPDATE %s SET'", s.buildTable()))
	}
	r, _ := regexp.Compile(`\[(\+|\-)\]?([a-zA-Z0-9_.\-\=\s\?\(\)]*)`)
	for i, vals := range s.params {
		if len(vals) == 0 {
			panic(fmt.Sprintf("Must be have values after 'UPDATE %s SET'", s.buildTable()))
		}
		if i == 0 {
			j := 0
			for arg, val := range vals {
				// fieldSlice = append(fieldSlice, arg)
				// s.render.args = append(s.render.args, val)
				if j > 0 {
					buffer.WriteString(", ")
				}

				match := r.FindStringSubmatch(arg)
				if len(match) > 1 {
					buffer.WriteString(buildIdent(match[2]))
					buffer.WriteString(" = ")
					buffer.WriteString(buildIdent(match[2]))
					buffer.WriteString(" ")
					buffer.WriteString(match[1])
					buffer.WriteString(" ?")
					s.render.args = append(s.render.args, val)
				} else {
					buffer.WriteString(buildIdent(arg))
					buffer.WriteString(" = ?")
					s.render.args = append(s.render.args, val)
				}
				j++
			}
		} else {
			//when update just support one of vals
			panic(fmt.Sprintf("when update just support one of vals: %v", vals))
		}
	}
	return buffer.String()
}

//Delete ...
func (s *SQLSegments) Delete() *SQLSegments {
	return s
}

//BuildDelete ...
func (s *SQLSegments) BuildDelete() string {
	var sql = fmt.Sprintf("DELETE%s FROM%s%s%s%s%s",
		s.buildFlags(),
		s.buildTable(),
		s.buildWhereClause(),
		s.buildOrderBy(),
		s.buildLimit(),
		s.buildReturning(),
	)
	s.cmd = _delete
	// fmt.Println(s.render.args)
	return sql
}

//buildIdent
func buildIdent(name string) string {
	return identKey + strings.Replace(name, ".", identKey+"."+identKey, -1) + identKey
}

func buildString(vals []string, sep string, header, footer string, ident bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(header)
	for i, s := range vals {
		if i > 0 {
			buffer.WriteString(sep)
		}
		if ident {
			buffer.WriteString(buildIdent(s))
		} else {
			buffer.WriteString(s)
		}
	}
	buffer.WriteString(footer)
	return buffer.String()
}

//buildPlaceholder
func buildPlaceholder(l int, holder, sep string) string {
	var buffer bytes.Buffer
	for i := 0; i < l; i++ {
		if i > 0 {
			buffer.WriteString(sep)
		}
		buffer.WriteString(holder)
	}
	return buffer.String()
}

//Args ..
func (s *SQLSegments) Args() []interface{} {
	return s.render.args
}

//Build ..
func (s *SQLSegments) Build() (string, []interface{}) {
	if s.cmd == _select {
		return s.BuildSelect(), s.Args()
	}
	if s.cmd == _insert {
		return s.BuildInsert(), s.Args()
	}
	if s.cmd == _replace {
		return s.BuildReplace(), s.Args()
	}
	if s.cmd == _update {
		return s.BuildUpdate(), s.Args()
	}
	if s.cmd == _delete {
		return s.BuildDelete(), s.Args()
	}
	return "", nil
}

//-------- another style --------

//Option ..
type Option func(q SQLSegments) SQLSegments

//Table ..
func Table(name interface{}) Option {
	return func(s SQLSegments) SQLSegments {
		s.Table(name)
		return s
	}
}

//Columns ..
func Columns(fields ...string) Option {
	return func(s SQLSegments) SQLSegments {
		s.Field(fields...)
		return s
	}
}

//Flag ..
func Flag(flags ...string) Option {
	return func(s SQLSegments) SQLSegments {
		s.Flag(flags...)
		return s
	}
}

//Join ..
func Join(table string, conditionA, logic, conditionB string) Option {
	return func(s SQLSegments) SQLSegments {
		s.Join(table, conditionA, logic, conditionB)
		return s
	}
}

//LeftJoin ..
func LeftJoin(table string, conditionA, logic, conditionB string) Option {
	// return func(s Query) Query {
	return func(s SQLSegments) SQLSegments {
		s.LeftJoin(table, conditionA, logic, conditionB)
		return s
	}
}

//RightJoin ..
func RightJoin(table string, conditionA, logic, conditionB string) Option {
	return func(s SQLSegments) SQLSegments {
		s.RightJoin(table, conditionA, logic, conditionB)
		return s
	}
}

//InnerJoin ..
func InnerJoin(table string, conditionA, logic, conditionB string) Option {
	return func(s SQLSegments) SQLSegments {
		s.InnerJoin(table, conditionA, logic, conditionB)
		return s
	}
}

//CorssJoin ..
func CorssJoin(table string, conditionA, logic, conditionB string) Option {
	return func(s SQLSegments) SQLSegments {
		s.CorssJoin(table, conditionA, logic, conditionB)
		return s
	}
}

//OrderBy ..
func OrderBy(fields ...string) Option {
	return func(s SQLSegments) SQLSegments {
		s.OrderBy(fields...)
		return s
	}
}

//GroupBy ..
func GroupBy(fields ...string) Option {
	return func(s SQLSegments) SQLSegments {
		s.GroupBy(fields...)
		return s
	}
}

//Offset ..
func Offset(n int) Option {
	return func(s SQLSegments) SQLSegments {
		s.Offset(n)
		return s
	}
}

//Limit ..
func Limit(n int) Option {
	return func(s SQLSegments) SQLSegments {
		s.Limit(n)
		return s
	}
}

//ForUpdate ..
func ForUpdate() Option {
	return func(s SQLSegments) SQLSegments {
		s.ForUpdate()
		return s
	}
}

//Returning ..
func Returning() Option {
	return func(s SQLSegments) SQLSegments {
		s.Returning()
		return s
	}
}

//Where ..
func Where(key interface{}, vals ...interface{}) Option {
	return func(s SQLSegments) SQLSegments {
		s.Where(key, vals...)
		return s
	}
}

//OrWhere ..
func OrWhere(key interface{}, vals ...interface{}) Option {
	return func(s SQLSegments) SQLSegments {
		s.OrWhere(key, vals...)
		return s
	}
}

//Set ..
func Set(key string, val interface{}) Option {
	//only use for update()
	return func(s SQLSegments) SQLSegments {
		s.UpdateField(key, val)
		return s
	}
}

//Params ...
func Params(vals ...map[string]interface{}) Option {
	return func(s SQLSegments) SQLSegments {
		s.Params(vals...)
		return s
	}
}

//BuildSQL ..
func buildSQL(cmd uint8, opts ...Option) (string, []interface{}) {
	s := SQLSegments{
		cmd: cmd,
	}
	for _, opt := range opts {
		s = opt(s)
	}
	return s.Build()
}

//SelectSQL ..
func SelectSQL(opts ...Option) (string, []interface{}) {
	return buildSQL(_select, opts...)
}

//InsertSQL ..
func InsertSQL(opts ...Option) (string, []interface{}) {
	return buildSQL(_insert, opts...)
}

//ReplaceSQL ..
func ReplaceSQL(opts ...Option) (string, []interface{}) {
	return buildSQL(_replace, opts...)
}

//UpdateSQL ..
func UpdateSQL(opts ...Option) (string, []interface{}) {
	return buildSQL(_update, opts...)
}

//DeleteSQL ..
func DeleteSQL(opts ...Option) (string, []interface{}) {
	return buildSQL(_delete, opts...)
}

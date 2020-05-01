package scanner

import (
	"fmt"
	"reflect"
	"testing"
)

func TestResolveModel1(t *testing.T) {
	type TestModel struct {
		ID int `db:"id,pk"`
	}
	m := TestModel{}
	ret, err := ResolveModelStruct(m)
	fmt.Println(ret, err)
}
func TestResolveModel2(t *testing.T) {
	type TestModel struct {
		ID int `db:"id,pk"`
	}
	m := &TestModel{}
	ret, err := ResolveModelStruct(m)
	fmt.Println(ret, err)
}
func TestResolveModel3(t *testing.T) {
	type TestModel struct {
		ID int `db:"id,pk"`
	}
	m := []*TestModel{}
	ret, err := ResolveModelStruct(m)
	fmt.Println(ret, err)
}
func TestResolveModel4(t *testing.T) {
	type TestModel struct {
		ID int `db:"id,pk"`
	}
	m := []TestModel{}
	ret, err := ResolveModelStruct(m)
	fmt.Println(ret, err)
}

func TestResolveModel5(t *testing.T) {
	type TestModel struct {
		ID   int    `db:"id,pk"`
		Name string `db:"column:name"`
		Age  string `db:"column:age,uni"`
	}
	m := TestModel{}
	m.ID = 999999
	m.Name = "jack"
	ret, _ := ResolveModelStruct(m)
	if ret.pk != "id" {
		t.Errorf("result: pk=%v, want: pk=%v", ret.pk, "id")
	}
	if ret.GetStructField("id").isPrimaryKey != true {
		t.Errorf("error: %v", `ret.GetStructField("id").isPrimaryKey`)
	}
	wantCols := []string{"id", "name", "age"}
	if !reflect.DeepEqual(ret.Columns(), wantCols) {
		t.Errorf("error: ret.Columns() result: %v want: %v", ret.Columns(), wantCols)
	}
	wantTbName := "TestModel"
	if ret.TableName() != wantTbName {
		t.Errorf("error: ret.TableName() result: %v want: %v", ret.TableName(), wantTbName)
	}
	wantPk := "id"
	if ret.GetPk() != wantPk {
		t.Errorf("error: ret.GetPk() result: %v want: %v", ret.GetPk(), wantPk)
	}

}

func TestResolveStructValue(t *testing.T) {
	type TestModel struct {
		ID int `db:"id"`
	}
	m := TestModel{}
	m.ID = 999999
	ret, _ := ResolveStructValue(m)
	result := fmt.Sprintf("%+v", ret)
	want := "map[id:999999]"
	if result != want {
		t.Errorf("result: %v, want: %v", result, want)
	}
}

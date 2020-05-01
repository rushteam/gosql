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
	if ret.GetStructField("id").column != "id" {
		t.Errorf("error: %v", `ret.GetStructField("id").column`)
	}
	if ret.GetStructField("name").column != "name" {
		t.Errorf("error: %v", `ret.GetStructField("id").column`)
	}
	if !reflect.DeepEqual(ret.Columns(), []string{"id", "name"}) {
		t.Errorf("error: %v", `ret.Columns()`)
	}
	if ret.TableName() != "TestModel" {
		t.Errorf("error: %v", `ret.TableName()`)
	}
	if ret.GetPk() != "id" {
		t.Errorf("error: %v", `ret.GetPk()`)
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

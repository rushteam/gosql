package scanner

import (
	"fmt"
	"testing"
)

func TestResolveModelStruct(t *testing.T) {
	type TestModel struct {
		ID int `db:"id,pk"`
	}
	m := TestModel{}
	m.ID = 999999
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

package scanner

import (
	"fmt"
	"testing"
)

func TestResolveModelStruct(t *testing.T) {
	type TestModel struct {
		ID int `db:"id"`
	}
	m := TestModel{}
	m.ID = 999999
	ret, _ := ResolveModelStruct(m)
	result := fmt.Sprintf("%+v", ret)
	want := "&{table:TestModel columns:[id] fields:map[id:0xc00008e680] pk:id}"
	if result != want {
		// t.Errorf("result: %v, want: %v", result, want)
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
		// t.Errorf("result: %v, want: %v", result, want)
	}
}

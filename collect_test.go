package gosql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestCollect1(t *testing.T) {
	NewCollect(nil, "test")
	clst := Collect("test")
	if clst != nil {
		t.Error("get cluster error want be nil")
	}
}

func TestCollect2(t *testing.T) {
	c := NewCluster()
	NewCollect(c, "test")
	clst := Collect("test")
	if clst == nil {
		t.Error("get cluster error,want be cluster")
	}
}

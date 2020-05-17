package gosql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	_ = mock
	defer db.Close()
	Debug = true
	clst := NewCluster(
		AddDb("mysql", ""),
	)
	s := &Session{ctx: context.TODO(), cluster: clst, v: 0}

	s.Master()
	s.Commit()
	s.Rollback()

	row := s.QueryRow("select 1")
	t.Log(row)

	rows, err := s.Query("select 1")
	if err != nil {
		t.Error(err)
	}
	t.Log(rows)

	rst, err := s.Exec("select 1")
	if err != nil {
		t.Error(err)
	}
	t.Log(rst)
}

package gosql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	_ = mock
	// defer db.Close()
	Debug = true
	columns := []string{"id"}
	mrows := mock.NewRows(columns).AddRow("1")
	mock.ExpectQuery("select * from test").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db}
	// s.Commit()
	// s.Rollback()

	row := s.QueryRow("select * from test")
	t.Log(row)

	// rows, err := s.Query("select * from test")
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Log(rows)

	// rst, err := s.Exec("select 1")
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Log(rst)
}

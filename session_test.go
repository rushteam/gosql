package gosql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSession(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// defer db.Close()
	columns := []string{"id"}
	mrows := mock.NewRows(columns).AddRow("100")
	mock.ExpectQuery("select * from test").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

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

	// s.Commit()
	// s.Rollback()
}

func TestSession1(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("INSERT INTO `tbl`").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	_, err = s.Exec("INSERT INTO `tbl` (`name`) values ('tom')")
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

type tblModel struct {
	Name string `db:"name"`
}

func (t1 *tblModel) TableName() string {
	return "tbl"
}

func TestSessionInsert(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("INSERT INTO `tbl`").WithArgs("川建国").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}
	t1 := &tblModel{}
	t1.Name = "川建国"
	_, err = s.Insert(t1)
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionReplace(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("REPLACE INTO `tbl`").WithArgs("川建国").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}
	t1 := &tblModel{}
	t1.Name = "川建国"
	_, err = s.Replace(t1)
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionDelete(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("DELETE FROM `tbl`").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}
	t1 := &tblModel{}
	_, err = s.Delete(t1)
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

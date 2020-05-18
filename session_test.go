package gosql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSession1(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	row := s.QueryRow("select * from test")
	t.Log(row)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	// s.Commit()
	// s.Rollback()
}

type tblModel struct {
	Name string `db:"name"`
}

func (t1 *tblModel) TableName() string {
	return "test"
}
func TestSessionExec(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("INSERT INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	_, err = s.Exec("INSERT INTO `test` (`name`) values ('tom')")
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionQuery1(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	row := s.QueryRow("select * from test")
	t.Log(row)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionQuery2(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	row := s.QueryRow("select * from test")
	t.Log(row)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionFetch(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("SELECT (.+) FROM `test`").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}
	t1 := &tblModel{}
	row := s.Fetch(t1)
	t.Log(row)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionFetchAll(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("SELECT (.+) FROM `test`").WillReturnRows(mrows)

	s := &Session{v: 0, executor: db, ctx: context.TODO()}
	t1 := &tblModel{}
	row := s.FetchAll(t1)
	t.Log(row)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSessionInsert(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO `test`").WithArgs("川建国").WillReturnResult(sqlmock.NewResult(2, 1))

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
	mock.ExpectExec("REPLACE INTO `test`").WithArgs("川建国").WillReturnResult(sqlmock.NewResult(2, 1))

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

func TestSessionUpdate(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE `test` SET").WithArgs("jerry", 1).WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	t1 := &tblModel{}
	t1.Name = "jerry"
	_, err = s.Update(t1, Where("id", 1))
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
	defer db.Close()
	mock.ExpectExec("DELETE FROM `test`").WillReturnResult(sqlmock.NewResult(2, 1))

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

package gosql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type t1Model struct {
	Name string `db:"name"`
}

func (t *t1Model) TableName() string {
	return "test"
}

type t2Model struct {
	ID   int64  `db:"id,pk"`
	Name string `db:"name"`
}

func (t *t2Model) TableName() string {
	return "test"
}

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
	t1 := &t1Model{}
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
	t1 := &t1Model{}
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

	t1 := &t1Model{}
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

	t1 := &t1Model{}
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

	t1 := &t1Model{}
	t1.Name = "jerry"
	_, err = s.Update(t1, Where("id", 1))
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSessionUpdate2(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE `test` SET").WithArgs("jerry", 1).WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	t1 := &t2Model{}
	t1.ID = 1
	t1.Name = "jerry"
	_, err = s.Update(t1)
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

	t1 := &t1Model{}
	_, err = s.Delete(t1)
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

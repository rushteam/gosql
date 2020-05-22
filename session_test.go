package gosql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSession1(t *testing.T) {
	//just err test
	Debug = true
	s := &Session{v: 0, executor: nil, ctx: context.TODO()}
	var err error
	_, err = s.Executor()
	t.Log(err)
	err = s.Rollback()
	t.Log(err)
	err = s.Commit()
	t.Log(err)

	_, err = s.Query("select 1")
	t.Log(err)
	_, err = s.QueryContext(context.Background(), "select 1")
	t.Log(err)

	_, err = s.Exec("set names utf8")
	t.Log(err)
	_, err = s.ExecContext(context.Background(), "set names utf8")
	t.Log(err)

	_, err = s.Insert(nil)
	t.Log(err)

	_, err = s.Update(nil)
	t.Log(err)

	_, err = s.Delete(nil)
	t.Log(err)

	err = s.Fetch(nil)
	t.Log(err)
	//fetchAll err
	err = s.FetchAll(nil)
	t.Log(err)
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
func TestSessionInsert2(t *testing.T) {
	AutoFillCreatedAtAndUpdatedAtField = true
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

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
	AutoFillCreatedAtAndUpdatedAtField = false
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
func TestSessionReplace2(t *testing.T) {
	AutoFillCreatedAtAndUpdatedAtField = true
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectExec("REPLACE INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

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
	AutoFillCreatedAtAndUpdatedAtField = false
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
func TestSessionUpdate3(t *testing.T) {
	AutoFillCreatedAtAndUpdatedAtField = true
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE `test` SET").WillReturnResult(sqlmock.NewResult(2, 1))

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
	AutoFillCreatedAtAndUpdatedAtField = false
}
func TestSessionDelete1(t *testing.T) {
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
func TestSessionDelete2(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("DELETE FROM `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	s := &Session{v: 0, executor: db, ctx: context.TODO()}

	t2 := &t2Model{}
	t2.ID = 1
	_, err = s.Delete(t2)
	if err != nil {
		t.Log(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestSessionTrans1(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()
	tx, _ := db.Begin()
	s := &Session{v: 0, executor: tx, ctx: context.TODO()}

	s.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSessionTrans2(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, _ := db.Begin()
	s := &Session{v: 0, executor: tx, ctx: context.TODO()}

	s.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

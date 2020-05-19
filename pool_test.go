package gosql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
)

func TestDbEngine(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	eng := &dbEngine{Db: db}
	eng.Connect()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster1(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	c := NewCluster()
	c.Exec("select 1")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster2(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	c := NewCluster(
		AddDb("mysql", "user:password@tcp(127.0.0.1:3306)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s", SetConnMaxLifetime(1), SetMaxIdleConns(1), SetMaxOpenConns(1)),
		AddDb("mysql", "user:password@tcp(127.0.0.1:3307)/test?parseTime=true&readTimeout=3s&writeTimeout=3s&timeout=3s", SetConnMaxLifetime(1), SetMaxIdleConns(1), SetMaxOpenConns(1)),
	)
	c.Query("select 1")
	c.Exec("select 2")
	s, _ := c.Master()
	s.Query("select 3")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func mockCluster(db *sql.DB) *PoolCluster {
	var pools []*dbEngine
	dbc := &dbEngine{
		Db:     db,
		Dsn:    "mockdb",
		Driver: "mysql",
	}
	pools = append(pools, dbc)
	c := &PoolCluster{}
	c.vs = 0
	c.pools = pools
	return c
}
func TestNewCluster3(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	c := mockCluster(db)
	rows, err := c.Query("select * from test")
	t.Log(rows, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestNewCluster4(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	c := mockCluster(db)
	rows, err := c.QueryContext(context.Background(), "select * from test")
	t.Log(rows, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster5(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	c := mockCluster(db)
	row := c.QueryRow("select * from test")
	t.Log(row)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster6(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mrows := sqlmock.NewRows([]string{"id", "name"}).AddRow("100", "tom")
	mock.ExpectQuery("select (.+) from test").WillReturnRows(mrows)

	c := mockCluster(db)
	row := c.QueryRowContext(context.Background(), "select * from test")
	t.Log(row)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster7(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	c := mockCluster(db)
	rst, err := c.Exec("INSERT INTO `test` (`name`) values ('tom')")
	t.Log(rst, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster8(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	c := mockCluster(db)
	rst, err := c.ExecContext(context.Background(), "INSERT INTO `test` (`name`) values ('tom')")
	t.Log(rst, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestNewCluster9(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	t1 := &t1Model{}
	t1.Name = "marry"
	c := mockCluster(db)
	rst, err := c.Insert(t1)
	t.Log(rst, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestNewCluster10(t *testing.T) {
	Debug = true

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("REPLACE INTO `test`").WillReturnResult(sqlmock.NewResult(2, 1))

	t2 := &t2Model{}
	t2.Name = "marry"
	c := mockCluster(db)
	rst, err := c.Replace(t2)
	t.Log(rst, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

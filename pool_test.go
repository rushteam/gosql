package gosql

import (
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

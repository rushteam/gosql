package gosql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
func TestNewCluster(t *testing.T) {
	Debug = true
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	c := NewCluster()
	c.Exec("select 1")
	t.Error(c)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

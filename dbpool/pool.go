package dbpool

import (
	"context"
	"database/sql"
)

//DB
type DB struct {
	Db *sql.DB
	ctx context.Context
}

func (o *Dbx) Query(string, ...interface{}) (sql.Result, error) {
	return o.Db.QueryContext(o.ctx, query, args...)
}

func (o *Dbx) Exec(string, ...interface{}) (sql.Result, error)
	return o.Db.ExecContext(o.ctx, query, args...)
}

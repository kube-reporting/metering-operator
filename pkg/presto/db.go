package presto

import "github.com/operator-framework/operator-metering/pkg/db"

type Queryer interface {
	Query(query string) ([]Row, error)
}

type Execer interface {
	Exec(query string) error
}

type ExecQueryer interface {
	Queryer
	Execer
}

type DB struct {
	queryer db.Queryer
}

func NewDB(queryer db.Queryer) *DB {
	return &DB{queryer}
}

func (db *DB) Query(query string) ([]Row, error) {
	return ExecuteSelect(db.queryer, query)
}

func (db *DB) Exec(query string) error {
	return ExecuteQuery(db.queryer, query)
}

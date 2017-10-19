package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type db struct {
	logger     log.FieldLogger
	logQueries bool

	db *sql.DB
}

func New(sqlDB *sql.DB, logger log.FieldLogger, logQueries bool) *db {
	return &db{
		db:         sqlDB,
		logger:     logger,
		logQueries: logQueries,
	}
}

func (db *db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if db.logQueries {
		margs := argsString(args...)
		db.logger.Debugf("QUERY: %s [%s]", query, margs)
	}
	return db.db.Query(query, args...)
}

// argsString pretty prints arguments passed into it for logging query
// arguments
func argsString(args ...interface{}) string {
	var margs string
	for i, a := range args {
		var v interface{} = a
		if x, ok := v.(driver.Valuer); ok {
			y, err := x.Value()
			if err == nil {
				v = y
			}
		}
		switch v.(type) {
		case string, []byte:
			v = fmt.Sprintf("%q", v)
		default:
			v = fmt.Sprintf("%v", v)
		}
		margs += fmt.Sprintf("%d:%s", i+1, v)
		if i+1 < len(args) {
			margs += " "
		}
	}
	return margs
}

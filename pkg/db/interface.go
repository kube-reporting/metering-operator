package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Close() error
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Close() error
}

type loggingQueryer struct {
	queryer    Queryer
	logger     log.FieldLogger
	logQueries bool
}

func NewLoggingQueryer(queryer Queryer, logger log.FieldLogger, logQueries bool) *loggingQueryer {
	return &loggingQueryer{
		queryer:    queryer,
		logger:     logger,
		logQueries: logQueries,
	}
}

func (loggingQueryer *loggingQueryer) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if loggingQueryer.logQueries {
		margs := argsString(args...)
		loggingQueryer.logger.Debugf("QUERY: %s [%s]", query, margs)
	}
	return loggingQueryer.queryer.Query(query, args...)
}

func (loggingQueryer *loggingQueryer) Close() error {
	return loggingQueryer.queryer.Close()
}

type loggingExecer struct {
	execer     Execer
	logger     log.FieldLogger
	logQueries bool
}

func NewLoggingExecer(execer Execer, logger log.FieldLogger, logQueries bool) *loggingExecer {
	return &loggingExecer{
		execer:     execer,
		logger:     logger,
		logQueries: logQueries,
	}
}

func (loggingExecer *loggingExecer) Exec(query string, args ...interface{}) (sql.Result, error) {
	if loggingExecer.logQueries {
		margs := argsString(args...)
		loggingExecer.logger.Debugf("EXEC: %s [%s]", query, margs)
	}
	return loggingExecer.execer.Exec(query, args...)
}

func (loggingExecer *loggingExecer) Close() error {
	return loggingExecer.execer.Close()
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

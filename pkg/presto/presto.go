package presto

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/avct/prestgo"
)

const (
	// TimestampFormat is the time format string used to produce Presto timestamps.
	TimestampFormat = "2006-01-02 15:04:05.000"
)

func prestoTime(t time.Time) string {
	return t.Format(TimestampFormat)
}

// ExecuteInsertQuery performs the query an INSERT into the table target. It's expected target has the correct schema.
func ExecuteInsertQuery(presto *sql.DB, target, query string) error {
	if presto == nil {
		return errors.New("presto instance of DB cannot be nil")
	}

	insert := fmt.Sprintf("INSERT INTO %s %s", target, query)
	result, err := presto.Query(insert)
	if err != nil {
		return err
	}
	for result.Next() {
		if err := result.Err(); err != nil {
			return fmt.Errorf("presto SQL error: %v", err)
		}
	}
	return nil
}

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
	rows, err := presto.Query(insert)
	if err != nil {
		return err
	}
	// Must call rows.Next() in order for errors to be populated correctly
	// because Query() only submits the query, and doesn't handle
	// success/failure. Next() is the method which inspects the submitted
	// queries status and causes errors to get stored in the sql.Rows object.
	for rows.Next() {
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("presto SQL error: %v", err)
	}
	return nil
}

// ExecuteSelectQuery performs the query on the table target. It's expected
// target has the correct schema.
func ExecuteSelect(prestoCon *sql.DB, query string) ([]map[string]interface{}, error) {
	if prestoCon == nil {
		return nil, errors.New("presto instance of DB cannot be nil")
	}

	rows, err := prestoCon.Query(query)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

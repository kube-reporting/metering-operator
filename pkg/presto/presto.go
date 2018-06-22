package presto

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/prestodb/presto-go-client/presto"

	"github.com/operator-framework/operator-metering/pkg/db"
)

const (
	// TimestampFormat is the time format string used to produce Presto timestamps.
	TimestampFormat = "2006-01-02 15:04:05.000"
)

func FormatInsertQuery(target, query string) string {
	return fmt.Sprintf("INSERT INTO %s %s", target, query)
}

// ExecuteInsertQuery performs the query an INSERT into the table target. It's expected target has the correct schema.
func ExecuteInsertQuery(queryer db.Queryer, target, query string) error {
	insertQuery := FormatInsertQuery(target, query)
	return ExecuteQuery(queryer, insertQuery)
}

func ExecuteQuery(queryer db.Queryer, query string) error {
	rows, err := queryer.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
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

type Row map[string]interface{}

// ExecuteSelectQuery performs the query on the table target. It's expected
// target has the correct schema.
func ExecuteSelect(queryer db.Queryer, query string) ([]Row, error) {
	rows, err := queryer.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []Row
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
		results = append(results, Row(m))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func DeleteFrom(prestoConn db.Queryer, tableName string) error {
	return ExecuteQuery(prestoConn, fmt.Sprintf("DELETE FROM %s", tableName))
}

type Column struct {
	Name string
	Type string
}

func GetRows(prestoConn db.Queryer, tableName string, columns []Column) ([]Row, error) {
	columnsSQL := GenerateQuotedColumnsListSQL(columns)
	orderBySQL := GenerateOrderBySQL(columns)
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", columnsSQL, tableName, orderBySQL)
	return ExecuteSelect(prestoConn, query)
}

func GenerateQuotedColumnsListSQL(columns []Column) string {
	var columnNames []string
	for _, col := range columns {
		columnNames = append(columnNames, quoteColumn(col))
	}
	columnsSQL := strings.Join(columnNames, ",")
	return columnsSQL
}

func GenerateOrderBySQL(columns []Column) string {
	var quotedColumns []string
	for _, col := range columns {
		colName := col.Name
		// if we detect a map(...) in the column, use map_entries to do
		// ordering. we detect a map column using a best effort approach by
		// checking if the column type contains the string "map(" , and is
		// followed by a ")" after that string.
		if mapIndex := strings.Index(col.Type, "map("); mapIndex != -1 && strings.Index(col.Type, ")") > mapIndex {
			quotedColumns = append(quotedColumns, fmt.Sprintf(`map_entries("%s")`, colName))
		} else {
			quotedColumns = append(quotedColumns, quoteColumn(col))
		}
	}
	return fmt.Sprintf("%s ASC", strings.Join(quotedColumns, ", "))
}

func Timestamp(date time.Time) string {
	return date.Format(TimestampFormat)
}

func quoteColumn(col Column) string {
	return `"` + col.Name + `"`
}

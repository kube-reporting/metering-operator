package presto

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// TimestampFormat is the time format string used to produce Presto timestamps.
	TimestampFormat = "2006-01-02 15:04:05.000"
)

func DeleteFrom(execer Execer, tableName string) error {
	return execer.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
}

func InsertInto(execer Execer, tableName, query string) error {
	return execer.Exec(FormatInsertQuery(tableName, query))
}

func GetRows(queryer Queryer, tableName string, columns []Column) ([]Row, error) {
	return queryer.Query(GenerateGetRowsSQL(tableName, columns))
}

func GenerateGetRowsSQL(tableName string, columns []Column) string {
	columnsSQL := GenerateQuotedColumnsListSQL(columns)
	orderBySQL := GenerateOrderBySQL(columns)
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", columnsSQL, tableName, orderBySQL)
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
		colType := strings.ToLower(col.Type)
		if mapIndex := strings.Index(colType, "map("); mapIndex != -1 && strings.Index(colType, ")") > mapIndex {
			quotedColumns = append(quotedColumns, fmt.Sprintf(`map_entries("%s")`, colName))
		} else {
			quotedColumns = append(quotedColumns, quoteColumn(col))
		}
	}
	return fmt.Sprintf("%s ASC", strings.Join(quotedColumns, ", "))
}

func FormatInsertQuery(target, query string) string {
	return fmt.Sprintf("INSERT INTO %s %s", target, query)
}

func Timestamp(input interface{}) (string, error) {
	var err error
	var d time.Time
	switch v := input.(type) {
	case time.Time:
		d = v
	case *time.Time:
		if v == nil {
			return "", errors.New("got nil timestamp")
		}
		d = *v
	case string:
		d, err = time.Parse(time.RFC3339, v)
	default:
		return "", fmt.Errorf("couldn't convert %#v to a Presto timestamp", input)
	}
	return d.Format(TimestampFormat), err
}

func quoteColumn(col Column) string {
	return `"` + col.Name + `"`
}

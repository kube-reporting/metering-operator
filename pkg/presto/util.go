package presto

import (
	"fmt"
	"strings"

	_ "github.com/prestodb/presto-go-client/presto"

	"github.com/operator-framework/operator-metering/pkg/db"
)

const (
	// TimestampFormat is the time format string used to produce Presto timestamps.
	TimestampFormat = "2006-01-02 15:04:05.000"
)

func DeleteFrom(queryer db.Queryer, tableName string) error {
	_, err := queryer.Query(fmt.Sprintf("DELETE FROM %s", tableName))
	return err
}

func InsertInto(queryer db.Queryer, tableName, query string) error {
	return execQuery(queryer, FormatInsertQuery(tableName, query))
}

func GetRows(queryer db.Queryer, tableName string, columns []Column) ([]Row, error) {
	return ExecuteSelect(queryer, GenerateGetRowsSQL(tableName, columns))
}

func GetRowsWhere(queryer db.Queryer, tableName string, columns []Column, whereClause string) ([]Row, error) {
	return ExecuteSelect(queryer, GenerateGetRowsSQLWithWhere(tableName, columns, whereClause))
}

func CreateTable(queryer db.Queryer, catalog, schema, tableName string, columns []Column, comment string, properties map[string]string, ignoreExists bool) error {
	query := generateCreateTableSQL(catalog, schema, tableName, columns, comment, properties, ignoreExists)
	_, err := queryer.Query(query)
	return err
}

func CreateTableAs(queryer db.Queryer, catalog, schema, tableName string, columns []Column, comment string, properties map[string]string, ignoreExists bool, query string) error {
	finalQuery := generateCreateTableAsSQL(catalog, schema, tableName, columns, comment, properties, ignoreExists, query)
	_, err := queryer.Query(finalQuery)
	return err
}

func DropTable(queryer db.Queryer, catalog, schema, tableName string, ignoreNotExists bool) error {
	ifExists := ""
	if ignoreNotExists {
		ifExists = "IF EXISTS"
	}
	table := FullyQualifiedTableName(catalog, schema, tableName)
	query := fmt.Sprintf("DROP TABLE %s %s", ifExists, table)
	_, err := queryer.Query(query)
	return err
}

func CreateView(queryer db.Queryer, catalog, schema, viewName string, query string, replace bool) error {
	fullQuery := "CREATE"
	if replace {
		fullQuery += " OR REPLACE"
	}
	fullQuery += " VIEW %s AS %s"
	view := FullyQualifiedTableName(catalog, schema, viewName)
	finalQuery := fmt.Sprintf(fullQuery, view, query)
	_, err := queryer.Query(finalQuery)
	return err
}

func DropView(queryer db.Queryer, catalog, schema, viewName string, ignoreNotExists bool) error {
	ifExists := ""
	if ignoreNotExists {
		ifExists = "IF EXISTS"
	}
	view := FullyQualifiedTableName(catalog, schema, viewName)
	query := fmt.Sprintf("DROP VIEW %s %s", ifExists, view)
	_, err := queryer.Query(query)
	return err
}

// QueryMetadata executes a "DESCRIBE" Presto query against an existing, fully-qualified
// table name to determine that table's column information.
func QueryMetadata(queryer db.Queryer, catalog, schema, tableName string) ([]Column, error) {
	rows, err := ExecuteSelect(queryer, fmt.Sprintf("DESCRIBE %s", FullyQualifiedTableName(catalog, schema, tableName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query the %s Presto table's metadata: %v", tableName, err)
	}

	var cols []Column
	for _, row := range rows {
		// note: row is in the form of map[string]interface{}
		// so we need to use type assertions to access values
		colName, ok := row["Column"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert the Presto column name to a string")
		}
		colType, ok := row["Type"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert the Presto column type to a string")
		}

		cols = append(cols, Column{
			Name: colName,
			Type: colType,
		})
	}

	return cols, nil
}

func GenerateGetRowsSQL(tableName string, columns []Column) string {
	columnsSQL := GenerateQuotedColumnsListSQL(columns)
	orderBySQL := GenerateOrderBySQL(columns)
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", columnsSQL, tableName, orderBySQL)
}

func GenerateGetRowsSQLWithWhere(tableName string, columns []Column, whereClause string) string {
	columnsSQL := GenerateQuotedColumnsListSQL(columns)
	orderBySQL := GenerateOrderBySQL(columns)
	return fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s", columnsSQL, tableName, whereClause, orderBySQL)
}

func GenerateQuotedColumnsListSQL(columns []Column) string {
	var columnNames []string
	for _, col := range columns {
		columnNames = append(columnNames, quoteColumn(col))
	}
	columnsSQL := strings.Join(columnNames, ",")
	return columnsSQL
}

func generateColumnDefinitionListSQL(columns []Column) string {
	c := make([]string, len(columns))
	for i, col := range columns {
		c[i] = fmt.Sprintf("`%s` %s", col.Name, col.Type)
	}
	return strings.Join(c, ",")
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

func FullyQualifiedTableName(catalog, schema, tableName string) string {
	return fmt.Sprintf("%s.%s.%s", catalog, schema, tableName)
}

func generateCreateTableSQL(catalog, schema, tableName string, columns []Column, comment string, properties map[string]string, ignoreExists bool) string {
	ifNotExists := ""
	if ignoreExists {
		ifNotExists = "IF NOT EXISTS"
	}

	columnsStr := generateColumnDefinitionListSQL(columns)

	commentStr := ""
	if comment != "" {
		commentStr = "COMMENT " + comment
	}

	propsStr := ""
	if len(properties) != 0 {
		propsStr = fmt.Sprintf("WITH (%s)", generatePropertiesSQL(properties))
	}

	table := FullyQualifiedTableName(catalog, schema, tableName)

	sqlStr := `CREATE TABLE %s
%s (
	%s
)
%s
%s`
	return fmt.Sprintf(sqlStr, ifNotExists, table, columnsStr, commentStr, propsStr)
}

func generateCreateTableAsSQL(catalog, schema, tableName string, columns []Column, comment string, properties map[string]string, ignoreExists bool, query string) string {
	ifNotExists := ""
	if ignoreExists {
		ifNotExists = "IF NOT EXISTS"
	}

	columnsStr := ""
	if columns != nil {
		columnsStr = GenerateQuotedColumnsListSQL(columns)
	}

	propsStr := ""
	if len(properties) != 0 {
		propsStr = fmt.Sprintf("WITH (%s)", generatePropertiesSQL(properties))
	}

	table := FullyQualifiedTableName(catalog, schema, tableName)

	sqlStr := `CREATE TABLE %s
%s (
	%s
)
%s
%s
AS %s`
	return fmt.Sprintf(sqlStr, ifNotExists, table, columnsStr, propsStr, comment, query)
}

func generatePropertiesSQL(props map[string]string) (propsTxt string) {
	var propList []string
	for k, v := range props {
		propList = append(propList, fmt.Sprintf("%s = %s", k, v))
	}
	return strings.Join(propList, ", ")
}

func FormatInsertQuery(target, query string) string {
	return fmt.Sprintf("INSERT INTO %s %s", target, query)
}

func quoteColumn(col Column) string {
	return `"` + col.Name + `"`
}

type Row map[string]interface{}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PartitionSpec map[string]string

type TablePartition struct {
	Location      string        `json:"location"`
	PartitionSpec PartitionSpec `json:"partitionSpec"`
}

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

func execQuery(queryer db.Queryer, query string) error {
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

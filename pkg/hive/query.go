package hive

import (
	"fmt"
	"strings"
)

func generateDropTableSQL(database, tableName string, ignoreNotExists, purge bool) string {
	ifExists := ""
	if ignoreNotExists {
		ifExists = "IF EXISTS"
	}
	purgeStr := ""
	if purge {
		purgeStr = "PURGE"
	}

	if database != "" {
		tableName = fmt.Sprintf("%s.%s", database, tableName)
	}

	return fmt.Sprintf("DROP TABLE %s %s %s", ifExists, tableName, purgeStr)
}

// generateCreateTableSQL returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func generateCreateTableSQL(params TableParameters, ignoreExists bool) string {
	columnsStr := generateColumnListSQL(params.Columns)

	tableType := ""
	if params.External {
		tableType = "EXTERNAL"
	}

	ifNotExists := ""
	if ignoreExists {
		ifNotExists = "IF NOT EXISTS"
	}

	tableName := params.Name
	if params.Database != "" {
		tableName = fmt.Sprintf("%s.%s", params.Database, tableName)
	}

	partitionedBy := ""
	if len(params.PartitionedBy) != 0 {
		partitionedBy = fmt.Sprintf("PARTITIONED BY (%s)", generateColumnListSQL(params.PartitionedBy))
	}

	clusteredBy := ""
	intoBuckets := ""
	sortedBy := ""
	if len(params.ClusteredBy) != 0 {
		clusteredBy = fmt.Sprintf("CLUSTERED BY (%s)", generateColumnNoTypesListSQL(params.ClusteredBy))
		intoBuckets = fmt.Sprintf("INTO %d BUCKETS", params.NumBuckets)

		if len(params.SortedBy) != 0 {
			sortedBy = fmt.Sprintf("SORTED BY (%s)", generateSortedByColumnListSQL(params.SortedBy))
		}
	}

	serdeFormatStr := ""
	if params.SerdeFormat != "" && params.SerdeRowProperties != nil {
		serdeFormatStr = fmt.Sprintf("ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s)", params.SerdeFormat, generateProperties(params.SerdeRowProperties))
	}
	format := ""
	if params.FileFormat != "" {
		format = fmt.Sprintf("STORED AS %s", params.FileFormat)
	}
	location := ""
	if params.Location != "" {
		location = fmt.Sprintf(`LOCATION "%s"`, params.Location)
	}
	tblProps := ""
	if params.Properties != nil {
		tblProps = fmt.Sprintf("TBLPROPERTIES(%s)", params.Properties)
	}
	return fmt.Sprintf(
		`CREATE %s TABLE %s
%s (%s) %s
%s %s %s
%s %s %s
%s`,
		tableType, ifNotExists,
		tableName, columnsStr, partitionedBy,
		clusteredBy, sortedBy, intoBuckets,
		serdeFormatStr, format, location,
		tblProps,
	)
}

func generateCreateDatabaseSQL(params DatabaseParameters, ignoreExists bool) string {
	ignoreExistsStr := ""
	if params.Location != "" {
		ignoreExistsStr = "IF NOT EXISTS"
	}
	locStr := ""
	if params.Location != "" {
		locStr = fmt.Sprintf("LOCATION '%s'", params.Location)
	}
	return fmt.Sprintf(
		`CREATE DATABASE
%s
%s
%s`, ignoreExistsStr, params.Name, locStr)
}

func generateDropDatabaseSQL(dbName string, ignoreNotExists, cascade bool) string {
	ifExists := ""
	if ignoreNotExists {
		ifExists = "IF EXISTS"
	}
	cascadeStr := ""
	if cascade {
		cascadeStr = "CASCADE"
	}
	return fmt.Sprintf(
		`DROP DATABASE
%s
%s
%s`, dbName, ifExists, cascadeStr)
}

// generateColumnListSQL returns a Hive CREATE column string from a slice of
// name/type pairs. For example, "columnName string".
func generateColumnListSQL(columns []Column) string {
	c := make([]string, len(columns))
	for i, col := range columns {
		c[i] = escapeColumn(col.Name, col.Type)
	}
	return strings.Join(c, ",")
}

func generateColumnNoTypesListSQL(columns []string) string {
	c := make([]string, len(columns))
	for i, col := range columns {
		c[i] = "`" + col + "`"
	}
	return strings.Join(c, ",")
}

func generateSortedByColumnListSQL(columns []SortColumn) string {
	c := make([]string, len(columns))
	for i, col := range columns {
		val := "`" + col.Name + "`"
		if col.Decending != nil {
			if *col.Decending {
				val += " DESC"
			} else {
				val += " ASC"
			}
		}
		c[i] = val
	}
	return strings.Join(c, ",")
}

func escapeColumn(columnName, columnType string) string {
	return fmt.Sprintf("`%s` %s", columnName, columnType)
}

// generateProperties returns a formatted a set of SerDe properties for a Hive query.
func generateProperties(props map[string]string) string {
	var propList []string
	for k, v := range props {
		propList = append(propList, fmt.Sprintf(`'%s' = '%s'`, k, v))
	}
	return strings.Join(propList, ", ")
}

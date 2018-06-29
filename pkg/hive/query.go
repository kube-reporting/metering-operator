package hive

import (
	"fmt"
	"strings"
)

func generateDropTableSQL(name string, ignoreNotExists, purge bool) string {
	ifExists := ""
	if ignoreNotExists {
		ifExists = "IF EXISTS"
	}
	purgeStr := ""
	if purge {
		purgeStr = "PURGE"
	}

	return fmt.Sprintf("DROP TABLE %s %s %s", ifExists, name, purgeStr)
}

// generateCreateTableSQL returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func generateCreateTableSQL(params TableParameters, properties TableProperties) string {
	columnsStr := generateColumnListSQL(params.Columns)

	tableType := ""
	if properties.External {
		tableType = "EXTERNAL"
	}

	ifNotExists := ""
	if params.IgnoreExists {
		ifNotExists = "IF NOT EXISTS"
	}

	partitionedBy := ""
	if len(params.Partitions) != 0 {
		partitionedBy = fmt.Sprintf("PARTITIONED BY (%s)", generateColumnListSQL(params.Partitions))
	}

	serdeFormatStr := ""
	if properties.SerdeFormat != "" && properties.SerdeProperties != nil {
		serdeFormatStr = fmt.Sprintf("ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s)", properties.SerdeFormat, generateSerdePropertiesSQL(properties.SerdeProperties))
	}
	location := ""
	if properties.Location != "" {
		location = fmt.Sprintf(`LOCATION "%s"`, properties.Location)
	}
	format := ""
	if properties.Format != "" {
		format = fmt.Sprintf("STORED AS %s", properties.Format)
	}
	return fmt.Sprintf(
		`CREATE %s TABLE %s
%s (%s) %s
%s %s %s`,
		tableType, ifNotExists,
		params.Name, columnsStr, partitionedBy,
		serdeFormatStr, format, location,
	)
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

func escapeColumn(columnName, columnType string) string {
	return fmt.Sprintf("`%s` %s", columnName, columnType)
}

// generateSerdePropertiesSQL returns a formatted a set of SerDe properties for a Hive query.
func generateSerdePropertiesSQL(props map[string]string) (propsTxt string) {
	first := true
	for k, v := range props {
		if !first {
			propsTxt += ", "
		}
		first = false

		pairStr := fmt.Sprintf("%q = %q", k, v)
		propsTxt += pairStr
	}
	return
}

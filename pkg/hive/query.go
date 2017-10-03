package hive

import (
	"fmt"
)

// createTable returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func createTable(name, location, serde string, serdeProps map[string]string, columns []string, external bool) string {
	serdePropsStr := fmtSerdeProps(serdeProps)
	columnsStr := fmtColumnText(columns)

	tableType := ""
	if external {
		tableType = "EXTERNAL"
	}

	query := "CREATE %s TABLE %s (%s) ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s) LOCATION \"%s\""
	return fmt.Sprintf(query, tableType, name, columnsStr, serde, serdePropsStr, location)
}

// fmtSerdeProps returns a formatted a set of SerDe properties for a Hive query.
func fmtSerdeProps(props map[string]string) (propsTxt string) {
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

// fmtColumnText returns a Hive CREATE column string from a slice of name/type pairs. For example, "columnName string".
func fmtColumnText(columns []string) (colTxt string) {
	for i, col := range columns {
		if i != 0 {
			colTxt += ", "
		}
		colTxt += col
	}
	return
}

// s3Location returns the HDFS path based on an S3 bucket and prefix.
func s3Location(bucket, prefix string) string {
	return fmt.Sprintf("s3a://%s/%s", bucket, prefix)
}

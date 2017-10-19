package hive

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func dropTable(name string, ignoreNotExists, purge bool) string {
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

// createTable returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func createTable(name, location, serdeFmt string, serdeProps map[string]string, columns []string, partitions map[string]string, external, ignoreExists bool) string {
	serdePropsStr := fmtSerdeProps(serdeProps)
	columnsStr := fmtColumnText(columns)

	tableType := ""
	if external {
		tableType = "EXTERNAL"
	}
	ifNotExists := ""
	if ignoreExists {
		ifNotExists = "IF NOT EXISTS"
	}
	partitionedBy := ""
	if partitions != nil {
		partitionedBy = fmt.Sprintf("PARTITIONED BY (%s)", fmtPartitionColText(partitions))
	}
	return fmt.Sprintf(
		`
CREATE %s TABLE %s
%s (%s) %s
ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s) LOCATION "%s"`,
		tableType, ifNotExists,
		name, columnsStr, partitionedBy,
		serdeFmt, serdePropsStr, location,
	)
}

func fmtPartitionColText(columns map[string]string) string {
	var c []string
	for columnName, columnType := range columns {
		c = append(c, fmt.Sprintf("`%s` %s", columnName, columnType))
	}
	return strings.Join(c, ",")
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
func s3Location(bucket, prefix string) (string, error) {
	bucket = path.Join(bucket, prefix)
	// Ensure the bucket URL has a trailing slash
	if bucket[len(bucket)-1] != '/' {
		bucket = bucket + "/"
	}
	location := "s3a://" + bucket

	locationURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return locationURL.String(), nil
}

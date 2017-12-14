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

type CreateTableParameters struct {
	Name         string
	Location     string
	SerdeFmt     string
	Format       string
	SerdeProps   map[string]string
	Columns      []Column
	Partitions   []Column
	External     bool
	IgnoreExists bool
}

// createTable returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func createTable(params CreateTableParameters) string {
	columnsStr := fmtColumnText(params.Columns)

	tableType := ""
	if params.External {
		tableType = "EXTERNAL"
	}

	ifNotExists := ""
	if params.IgnoreExists {
		ifNotExists = "IF NOT EXISTS"
	}

	partitionedBy := ""
	if len(params.Partitions) != 0 {
		partitionedBy = fmt.Sprintf("PARTITIONED BY (%s)", fmtColumnText(params.Partitions))
	}

	serdeFormatStr := ""
	if params.SerdeFmt != "" && params.SerdeProps != nil {
		serdeFormatStr = fmt.Sprintf("ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s)", params.SerdeFmt, fmtSerdeProps(params.SerdeProps))
	}
	location := ""
	if params.Location != "" {
		location = fmt.Sprintf(`LOCATION "%s"`, params.Location)
	}
	format := ""
	if params.Format != "" {
		format = fmt.Sprintf("STORED AS %s", params.Format)
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

type Column struct {
	Name string
	Type string
}

func fmtColumnText(columns []Column) string {
	c := make([]string, len(columns))
	for i, col := range columns {
		c[i] = escapeColumn(col.Name, col.Type)
	}
	return strings.Join(c, ",")
}

func escapeColumn(columnName, columnType string) string {
	return fmt.Sprintf("`%s` %s", columnName, columnType)
}

// fmtSerdeProps returns a formatted a set of SerDe properties for a Hive query.
func fmtSerdeProps(props map[string]string) (propsTxt string) {
	***REMOVED***rst := true
	for k, v := range props {
		if !***REMOVED***rst {
			propsTxt += ", "
		}
		***REMOVED***rst = false

		pairStr := fmt.Sprintf("%q = %q", k, v)
		propsTxt += pairStr
	}
	return
}

// fmtColumnText returns a Hive CREATE column string from a slice of name/type pairs. For example, "columnName string".
// s3Location returns the HDFS path based on an S3 bucket and pre***REMOVED***x.
func s3Location(bucket, pre***REMOVED***x string) (string, error) {
	bucket = path.Join(bucket, pre***REMOVED***x)
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

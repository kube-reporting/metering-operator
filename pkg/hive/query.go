package hive

import (
	"fmt"
	"net/url"
	"path"
)

// createTable returns a query for a CREATE statement which instantiates a new external Hive table.
// If is external is set, an external Hive table will be used.
func createTable(name, location, serdeFmt string, serdeProps map[string]string, columns []string, external, ignoreExists bool) string {
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
	return fmt.Sprintf(
		`
CREATE %s TABLE %s
%s ( %s)
ROW FORMAT SERDE '%s' WITH SERDEPROPERTIES (%s) LOCATION "%s"`,
		tableType, ifNotExists,
		name, columnsStr,
		serdeFmt, serdePropsStr, location,
	)
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
func fmtColumnText(columns []string) (colTxt string) {
	for i, col := range columns {
		if i != 0 {
			colTxt += ", "
		}
		colTxt += col
	}
	return
}

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

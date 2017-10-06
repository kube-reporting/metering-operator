package hive

import "errors"

// CreateReportTable creates a new table backed by the given bucket/prefix with
// the specified columns
func CreateReportTable(conn *Connection, tableName, bucket, prefix string, columns []string) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} else if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}
	query := createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, false, false)
	return conn.Query(query)
}

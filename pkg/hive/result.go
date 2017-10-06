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

	// use s3n HDFS driver for s3
	location := s3Location(bucket, prefix)
	query := createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, false)
	return conn.Query(query)
}

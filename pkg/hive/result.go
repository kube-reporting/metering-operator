package hive

import (
	"errors"
	"fmt"
)

// CreateReportTable creates a new table backed by the given bucket/pre***REMOVED***x with
// the speci***REMOVED***ed columns
func CreateReportTable(conn *Connection, tableName, bucket, pre***REMOVED***x string, columns []string) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} ***REMOVED*** if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	location, err := s3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return err
	}

	query := dropTable(tableName, true)
	err = conn.Query(query)
	if err != nil {
		return err
	}

	query = createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, false, true)
	return conn.Query(query)
}

func ExecuteTruncate(conn *Connection, tableName string) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} ***REMOVED*** if conn.session == nil {
		return errors.New("the Hive session has closed")
	}
	return conn.Query(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
}

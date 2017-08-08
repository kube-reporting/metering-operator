package hive

import "errors"

var (
	PodCostColumns = []string{
		"pod string",
		"namespace string",
		"node string",
		"cost double",
		"begin timestamp",
		"stop timestamp",
		"labels map<string, string>",
	}

	PodUsageColumns = []string{
		"pod string",
		"namespace string",
		"node string",
		"usage double",
		"begin timestamp",
		"stop timestamp",
		"labels map<string, string>",
	}
)

// CreatePodCostTable instantiates a new Hive table to hold the result of a Pod/dollar report.
func CreatePodCostTable(conn *Connection, tableName, bucket, prefix string) error {
	return createReportTable(conn, tableName, bucket, prefix, PodCostColumns)
}

// CreatePodUsageTable instantiates a table for Pod usage aggregates.
func CreatePodUsageTable(conn *Connection, tableName, bucket, prefix string) error {
	return createReportTable(conn, tableName, bucket, prefix, PodUsageColumns)
}

func createReportTable(conn *Connection, tableName, bucket, prefix string, columns []string) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} else if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	// use s3n HDFS driver for s3
	location := s3nLocation(bucket, prefix)
	query := createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, false)
	return conn.Query(query)
}

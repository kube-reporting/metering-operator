package hive

import (
	"fmt"
)

// CreateReportTable creates a new table backed by the given bucket/prefix with
// the specified columns
func CreateReportTable(queryer Queryer, tableName, bucket, prefix string, columns []string) error {
	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}

	query := dropTable(tableName, true)
	err = queryer.Query(query)
	if err != nil {
		return err
	}

	query = createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, false, true)
	return queryer.Query(query)
}

func ExecuteTruncate(queryer Queryer, tableName string) error {
	return queryer.Query(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
}

package hive

import (
	"path"
)

var (
	promsumColumns = []Column{
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map<string, string>"},
	}
)

// CreatePromsumTable instantiates a new Hive table for Prometheus observation
// data stored in S3.
func CreatePromsumTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string) error {
	path := path.Join(pre***REMOVED***x, tableName)
	location, err := s3Location(bucket, path)
	if err != nil {
		return err
	}
	query := createTable(tableName, location, "", "", nil, promsumColumns, nil, false, true)
	return queryer.Query(query)
}

// CreateLocalPromsumTable instantiates a new Hive table for Prometheus
// observation data stored locally
func CreateLocalPromsumTable(queryer Queryer, tableName string) error {
	query := createTable(tableName, "", "", "", nil, promsumColumns, nil, false, true)
	return queryer.Query(query)
}

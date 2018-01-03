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
func CreateS3PromsumTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string) (CreateTableParameters, error) {
	path := path.Join(pre***REMOVED***x, tableName)
	location, err := S3Location(bucket, path)
	if err != nil {
		return CreateTableParameters{}, err
	}

	params := CreateTableParameters{
		Name:         tableName,
		Location:     location,
		SerdeFmt:     "",
		Format:       "",
		SerdeProps:   nil,
		Columns:      promsumColumns,
		Partitions:   nil,
		External:     false,
		IgnoreExists: true,
	}
	query := createTable(params)
	return params, queryer.Query(query)
}

// CreateLocalPromsumTable instantiates a new Hive table for Prometheus
// observation data stored locally
func CreateLocalPromsumTable(queryer Queryer, tableName string) (CreateTableParameters, error) {
	params := CreateTableParameters{
		Name:         tableName,
		Location:     "",
		SerdeFmt:     "",
		Format:       "",
		SerdeProps:   nil,
		Columns:      promsumColumns,
		Partitions:   nil,
		External:     false,
		IgnoreExists: true,
	}
	query := createTable(params)
	return params, queryer.Query(query)
}

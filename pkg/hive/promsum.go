package hive

var (
	promsumColumns = []Column{
		{Name: "query", Type: "string"},
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map<string, string>"},
	}
)

// CreatePromsumTable instantiates a new external Hive table for Prometheus observation data stored in S3.
func CreatePromsumTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string) error {
	location, err := s3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return err
	}
	query := createTable(tableName, location, "", nil, promsumColumns, nil, false, true)
	return queryer.Query(query)
}

// CreateLocalPromsumTable instantiates a new external Hive table for Prometheus
// observation data stored locally
func CreateLocalPromsumTable(queryer Queryer, tableName string) error {
	query := createTable(tableName, "", "", nil, promsumColumns, nil, false, true)
	return queryer.Query(query)
}

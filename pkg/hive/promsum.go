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
func CreatePromsumTable(queryer Queryer, tableName, bucket, prefix string) error {
	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}
	query := createTable(tableName, location, "", nil, promsumColumns, nil, false, true)
	return queryer.Query(query)
}

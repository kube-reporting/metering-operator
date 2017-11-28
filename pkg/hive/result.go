package hive

// CreateReportTable creates a new table backed by the given bucket/pre***REMOVED***x with
// the speci***REMOVED***ed columns
func CreateReportTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string, columns []Column) error {
	location, err := s3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return err
	}

	query := dropTable(tableName, true, true)
	err = queryer.Query(query)
	if err != nil {
		return err
	}

	query = createTable(tableName, location, "", "", nil, columns, nil, false, false)
	return queryer.Query(query)
}

func CreateLocalReportTable(queryer Queryer, tableName string, columns []Column) error {
	query := dropTable(tableName, true, true)
	err := queryer.Query(query)
	if err != nil {
		return err
	}

	query = createTable(tableName, "", "", "", nil, columns, nil, false, true)
	return queryer.Query(query)
}

package hive

import "path"

// CreateReportTable creates a new table backed by the given bucket/pre***REMOVED***x with
// the speci***REMOVED***ed columns
func CreateS3ReportTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string, columns []Column) error {
	path := path.Join(pre***REMOVED***x, tableName)
	location, err := s3Location(bucket, path)
	if err != nil {
		return err
	}

	query := dropTable(tableName, true, true)
	err = queryer.Query(query)
	if err != nil {
		return err
	}

	query = createTable(CreateTableParameters{tableName, location, "", "", nil, columns, nil, false, false})
	return queryer.Query(query)
}

func CreateLocalReportTable(queryer Queryer, tableName string, columns []Column) error {
	query := dropTable(tableName, true, true)
	err := queryer.Query(query)
	if err != nil {
		return err
	}

	query = createTable(CreateTableParameters{tableName, "", "", "", nil, columns, nil, false, true})
	return queryer.Query(query)
}

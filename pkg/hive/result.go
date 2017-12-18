package hive

import "path"

// CreateReportTable creates a new table backed by the given bucket/prefix with
// the specified columns
func CreateS3ReportTable(queryer Queryer, tableName, bucket, prefix string, columns []Column, drop bool) (CreateTableParameters, error) {
	path := path.Join(prefix, tableName)
	location, err := S3Location(bucket, path)
	if err != nil {
		return CreateTableParameters{}, err
	}

	if drop {
		err := DropTable(queryer, tableName, true)
		if err != nil {
			return CreateTableParameters{}, err
		}
	}
	params := CreateTableParameters{
		Name:         tableName,
		Location:     location,
		SerdeFmt:     "",
		Format:       "",
		SerdeProps:   nil,
		Columns:      columns,
		Partitions:   nil,
		External:     false,
		IgnoreExists: false,
	}
	query := createTable(params)
	return params, queryer.Query(query)
}

func CreateLocalReportTable(queryer Queryer, tableName string, columns []Column, drop bool) (CreateTableParameters, error) {
	if drop {
		err := DropTable(queryer, tableName, true)
		if err != nil {
			return CreateTableParameters{}, err
		}
	}

	params := CreateTableParameters{
		Name:         tableName,
		Location:     "",
		SerdeFmt:     "",
		Format:       "",
		SerdeProps:   nil,
		Columns:      columns,
		Partitions:   nil,
		External:     false,
		IgnoreExists: true,
	}
	query := createTable(params)
	return params, queryer.Query(query)
}

func DropTable(queryer Queryer, tableName string, ignoreNotExists bool) error {
	query := dropTable(tableName, ignoreNotExists, true)
	return queryer.Query(query)
}

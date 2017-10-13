package hive

var (
	// PromsumTableName is the Hive identifier to use for usage data.
	PromsumTableName = "promsum"

	// PromsumSerde specifies the Hadoop serialization/deserialization implementation to be used.
	PromsumSerde = "org.apache.hive.hcatalog.data.JsonSerDe"

	// PromsumSerdeProps define the behavior of the SerDe used with promsum data.
	PromsumSerdeProps = map[string]string{
		"timestamp.formats": "yyyy-MM-dd'T'HH:mm:ss.SSSZ",
	}

	PromsumColumns = []string{
		"query string",
		"amount double",
		"`timestamp` timestamp",
		"`timePrecision` double",
		"labels map<string, string>",
	}
)

// CreatePromsumTable instantiates a new external Hive table for Prometheus observation data stored in S3.
func CreatePromsumTable(queryer Queryer, tableName, bucket, prefix string) error {
	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}
	query := createTable(tableName, location, PromsumSerde, PromsumSerdeProps, PromsumColumns, nil, false, true)
	return queryer.Query(query)
}

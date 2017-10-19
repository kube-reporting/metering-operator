package hive

var (
	// PromsumSerde speci***REMOVED***es the Hadoop serialization/deserialization implementation to be used.
	PromsumSerde = "org.apache.hive.hcatalog.data.JsonSerDe"

	// PromsumSerdeProps de***REMOVED***ne the behavior of the SerDe used with promsum data.
	PromsumSerdeProps = map[string]string{
		"timestamp.formats": "yyyy-MM-dd'T'HH:mm:ss.SSSZ",
	}

	PromsumColumns = []Column{
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
	query := createTable(tableName, location, PromsumSerde, PromsumSerdeProps, PromsumColumns, nil, false, true)
	return queryer.Query(query)
}

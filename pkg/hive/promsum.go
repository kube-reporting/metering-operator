package hive

import (
	"errors"
)

var (
	// PromsumTableName is the Hive identi***REMOVED***er to use for usage data.
	PromsumTableName = "promsum"

	// PromsumSerde speci***REMOVED***es the Hadoop serialization/deserialization implementation to be used.
	PromsumSerde = "org.apache.hive.hcatalog.data.JsonSerDe"

	// PromsumSerdeProps de***REMOVED***ne the behavior of the SerDe used with promsum data.
	PromsumSerdeProps = map[string]string{
		"timestamp.formats": "yyyy-MM-dd'T'HH:mm:ss.SSSZ",
	}

	PromsumColumns = []string{
		"subject string",
		"amount float",
		"`start` timestamp",
		"`end` timestamp",
		"labels map<string, string>",
	}
)

// CreatePromsumTable instantiates a new external Hive table for Prometheus observation data stored in S3.
func CreatePromsumTable(conn *Connection, bucket, pre***REMOVED***x string) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} ***REMOVED*** if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	// use s3n HDFS driver for s3
	location := s3nLocation(bucket, pre***REMOVED***x)
	query := createExternalTbl(PromsumTableName, location, PromsumSerde, PromsumSerdeProps, PromsumColumns)
	return conn.Query(query)
}

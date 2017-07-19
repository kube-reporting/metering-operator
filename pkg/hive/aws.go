package hive

import (
	"errors"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
)

var (
	// AWSUsageTableName is the Hive identi***REMOVED***er to use for the AWS Billing Data table.
	AWSUsageTableName = "awsBilling"

	// AWSUsageSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	AWSUsageSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"

	// AWSUsageSerdeProps con***REMOVED***gure the SerDe used with AWS Billing Data.
	AWSUsageSerdeProps = map[string]string{
		"serialization.format": ",",
		"***REMOVED***eld.delim":          ",",
		"collection.delim":     "unde***REMOVED***ned",
		"mapkey.delim":         "unde***REMOVED***ned",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func CreateAWSUsageTable(conn *Connection, bucket string, manifest aws.Manifest) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} ***REMOVED*** if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	// TODO: support for multiple partitions
	location := s3nLocation(bucket, manifest.Paths()[0])
	columns := manifest.Columns.HQL()

	query := createExternalTbl(AWSUsageTableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns)
	return conn.Query(query)
}

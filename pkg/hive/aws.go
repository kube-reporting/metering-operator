package hive

import (
	"errors"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
)

var (
	// AWSUsageTableName is the Hive identifier to use for the AWS Billing Data table.
	AWSUsageTableName = "awsBilling"

	// AWSUsageSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	AWSUsageSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"

	// AWSUsageSerdeProps configure the SerDe used with AWS Billing Data.
	AWSUsageSerdeProps = map[string]string{
		"serialization.format": ",",
		"field.delim":          ",",
		"collection.delim":     "undefined",
		"mapkey.delim":         "undefined",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func CreateAWSUsageTable(conn *Connection, bucket string, manifest aws.Manifest) error {
	if conn == nil {
		return errors.New("connection to Hive cannot be nil")
	} else if conn.session == nil {
		return errors.New("the Hive session has closed")
	}

	// TODO: support for multiple partitions
	location := s3nLocation(bucket, manifest.Paths()[0])
	columns := manifest.Columns.HQL()

	query := createTable(AWSUsageTableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, true)
	return conn.Query(query)
}

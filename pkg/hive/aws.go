package hive

import (
	"errors"
	"fmt"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
)

var (
	// AWSUsageTableName is the Hive identifier to use for the AWS Billing Data table.
	AWSUsageTableName = "awsBilling"

	// AWSUsageSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	AWSUsageSerde = "org.apache.hive.hcatalog.data.JsonSerDe"

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

	columns := make([]string, len(manifest.Columns))
	for i, c := range manifest.Columns {
		columns[i] = fmt.Sprintf("%s %s", c.HiveName(), c.HiveType())
	}

	query := createExternalTbl(AWSUsageTableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns)
	return conn.Query(query)
}

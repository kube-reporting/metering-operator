package hive

import (
	"fmt"
	"strings"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
)

var (
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

	awsPartitions = map[string]string{
		"assemblyId":           "string",
		"billing_period_start": "timestamp",
		"billing_period_end":   "timestamp",
	}
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func CreateAWSUsageTable(queryer Queryer, tableName, bucket, prefix string, manifest *aws.Manifest) error {
	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}
	columns := manifest.Columns.HQL()

	query := createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, awsPartitions, true, true)
	return queryer.Query(query)
}

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

const hiveTimestampLayout = "2006-01-02 15:04:05.000000000"

func UpdateAWSUsageTable(queryer Queryer, tableName, bucket, prefix string, manifests []*aws.Manifest) error {
	partitionStr := "PARTITION (`billing_period_start`='%s', `billing_period_end`='%s', `assemblyId`='%s') LOCATION '%s'"
	var stmts []string
	// A map containing locations we've already added to ensure we do not
	// attempt to add a partition multiple times, in case the assemblyId shows
	// up in multiple manifests
	locations := make(map[string]struct{})
	for _, manifest := range manifests {
		for _, manifestPath := range manifest.Paths() {
			location, err := s3Location(bucket, manifestPath)
			if err != nil {
				return err
			}
			if _, exists := locations[location]; exists {
				continue
			}
			locations[location] = struct{}{}
			stmt := fmt.Sprintf(partitionStr, manifest.BillingPeriod.Start.Format(hiveTimestampLayout), manifest.BillingPeriod.End.Format(hiveTimestampLayout), manifest.AssemblyID, location)
			stmts = append(stmts, stmt)
		}

	}
	if len(stmts) == 0 {
		return nil
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD IF NOT EXISTS %s", tableName, strings.Join(stmts, " "))
	return queryer.Query(query)
}

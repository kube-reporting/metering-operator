package reportingutil

import (
	"fmt"
	"strings"

	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

const (
	// AWSUsageHiveSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	AWSUsageHiveSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"

	// AWSUsagePartitionDateStringLayout is the format used to partition
	// AWSUsage partition key
	AWSUsagePartitionDateStringLayout = "20060102"
)

var (
	// AWSUsageHiveSerdeProps configure the SerDe used with AWS Billing Data.
	AWSUsageHiveSerdeProps = map[string]string{
		"serialization.format": ",",
		"field.delim":          ",",
		"collection.delim":     "undefined",
		"mapkey.delim":         "undefined",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}

	AWSUsageHivePartitions = []hive.Column{
		{Name: "billing_period_start", Type: "string"},
		{Name: "billing_period_end", Type: "string"},
	}
)

// AddAWSHivePartition will add a new partition to the given tableName for the time
// range, pointing at the location
func AddAWSHivePartition(queryer db.Queryer, tableName, start, end, location string) error {
	partitionStr := "ALTER TABLE %s ADD IF NOT EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s') LOCATION '%s'"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end, location)
	_, err := queryer.Query(stmt)
	return err
}

// DropAWSHivePartition will delete a partition from the given tableName for the time
// range, pointing at the location
func DropAWSHivePartition(queryer db.Queryer, tableName, start, end string) error {
	partitionStr := "ALTER TABLE %s DROP IF EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s')"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end)
	_, err := queryer.Query(stmt)
	return err
}

// SanetizeAWSColumnForHive removes and replaces invalid characters in AWS
// billing columns with characters allowed in hive SQL
func SanetizeAWSColumnForHive(col aws.Column) string {
	name := fmt.Sprintf("%s_%s", strings.TrimSpace(col.Category), strings.TrimSpace(col.Name))
	// hive does not allow ':' or '.' in identifiers
	name = strings.Replace(name, ":", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return strings.ToLower(name)
}

// AWSColumnToHiveColumnType is the data type a column is created as in Hive.
func AWSColumnToHiveColumnType(c aws.Column) string {
	switch SanetizeAWSColumnForHive(c) {
	case "lineitem_usagestartdate", "lineitem_usageenddate":
		return "timestamp"
	case "lineitem_blendedcost":
		return "double"
	default:
		return "string"
	}
}

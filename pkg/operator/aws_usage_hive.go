package operator

import (
	"fmt"
	"strings"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/sirupsen/logrus"
)

var (
	// awsUsageHiveSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	awsUsageHiveSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"

	// awsUsageHiveSerdeProps configure the SerDe used with AWS Billing Data.
	awsUsageHiveSerdeProps = map[string]string{
		"serialization.format": ",",
		"field.delim":          ",",
		"collection.delim":     "undefined",
		"mapkey.delim":         "undefined",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}

	awsUsageHivePartitions = []hive.Column{
		{Name: "billing_period_start", Type: "string"},
		{Name: "billing_period_end", Type: "string"},
	}
)

const awsUsagePartitionDateStringLayout = "20060102"

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func (op *Reporting) createAWSUsageTable(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName, bucket, prefix string, manifests []*aws.Manifest) error {
	location, err := hive.S3Location(bucket, prefix)
	if err != nil {
		return err
	}

	// Since the billing data likely exists already, we need to enumerate all
	// columns for all manifests to get the entire set of columns used
	// historically.
	// TODO(chance): We will likely want to do this when we add partitions
	// to avoid having to do it all up front.
	columns := make([]hive.Column, 0)
	seen := make(map[string]struct{})
	for _, manifest := range manifests {
		for _, c := range manifest.Columns {
			name := sanetizeAWSColumnForHive(c)
			colType := awsColumnToHiveColumnType(c)

			if _, exists := seen[name]; !exists {
				seen[name] = struct{}{}
				columns = append(columns, hive.Column{
					Name: name,
					Type: colType,
				})
			}
		}
	}

	params := hive.TableParameters{
		Name:         tableName,
		Columns:      columns,
		Partitions:   awsUsageHivePartitions,
		IgnoreExists: true,
	}
	properties := hive.TableProperties{
		Location:           location,
		FileFormat:         "textfile",
		SerdeFormat:        awsUsageHiveSerde,
		SerdeRowProperties: awsUsageHiveSerdeProps,
		External:           true,
	}
	return op.createTableWith(logger, dataSource, "ReportDataSource", dataSource.Name, params, properties)
}

// addAWSHivePartition will add a new partition to the given tableName for the time
// range, pointing at the location
func addAWSHivePartition(queryer db.Queryer, tableName, start, end, location string) error {
	partitionStr := "ALTER TABLE %s ADD IF NOT EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s') LOCATION '%s'"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end, location)
	_, err := queryer.Query(stmt)
	return err
}

// dropAWSHivePartition will delete a partition from the given tableName for the time
// range, pointing at the location
func dropAWSHivePartition(queryer db.Queryer, tableName, start, end string) error {
	partitionStr := "ALTER TABLE %s DROP IF EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s')"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end)
	_, err := queryer.Query(stmt)
	return err
}

// sanetizeAWSColumnForHive removes and replaces invalid characters in AWS
// billing columns with characters allowed in hive SQL
func sanetizeAWSColumnForHive(col aws.Column) string {
	name := fmt.Sprintf("%s_%s", strings.TrimSpace(col.Category), strings.TrimSpace(col.Name))
	// hive does not allow ':' or '.' in identifiers
	name = strings.Replace(name, ":", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return strings.ToLower(name)
}

// awsColumnToHiveColumnType is the data type a column is created as in Hive.
func awsColumnToHiveColumnType(c aws.Column) string {
	switch sanetizeAWSColumnForHive(c) {
	case "lineitem_usagestartdate", "lineitem_usageenddate":
		return "timestamp"
	case "lineitem_blendedcost":
		return "double"
	default:
		return "string"
	}
}

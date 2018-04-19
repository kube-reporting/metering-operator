package hive

import (
	"fmt"
	"strings"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
)

var (
	// awsUsageSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	awsUsageSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"

	// awsUsageSerdeProps con***REMOVED***gure the SerDe used with AWS Billing Data.
	awsUsageSerdeProps = map[string]string{
		"serialization.format": ",",
		"***REMOVED***eld.delim":          ",",
		"collection.delim":     "unde***REMOVED***ned",
		"mapkey.delim":         "unde***REMOVED***ned",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}

	awsPartitions = []Column{
		{Name: "billing_period_start", Type: "string"},
		{Name: "billing_period_end", Type: "string"},
	}
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func CreateAWSUsageTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string, manifests []*aws.Manifest) (CreateTableParameters, error) {
	location, err := S3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return CreateTableParameters{}, err
	}

	// Since the billing data likely exists already, we need to enumerate all
	// columns for all manifests to get the entire set of columns used
	// historically.
	// TODO(chance): We will likely want to do this when we add partitions
	// to avoid having to do it all up front.
	columns := make([]Column, 0)
	seen := make(map[string]struct{})
	for _, manifest := range manifests {
		for _, c := range manifest.Columns {
			name := hiveName(c)
			colType := hiveType(c)

			if _, exists := seen[name]; !exists {
				seen[name] = struct{}{}
				columns = append(columns, Column{
					Name: name,
					Type: colType,
				})
			}
		}
	}

	params := CreateTableParameters{
		Name:         tableName,
		Location:     location,
		SerdeFmt:     awsUsageSerde,
		Format:       "text***REMOVED***le",
		SerdeProps:   awsUsageSerdeProps,
		Columns:      columns,
		Partitions:   awsPartitions,
		External:     true,
		IgnoreExists: true,
	}
	query := createTable(params)
	return params, queryer.Query(query)
}

const HiveDateStringLayout = "20060102"

// AddPartition will add a new partition to the given tableName for the time
// range, pointing at the location
func AddPartition(queryer Queryer, tableName, start, end, location string) error {
	partitionStr := "ALTER TABLE %s ADD IF NOT EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s') LOCATION '%s'"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end, location)
	return queryer.Query(stmt)
}

// DropPartition will delete a partition from the given tableName for the time
// range, pointing at the location
func DropPartition(queryer Queryer, tableName, start, end string) error {
	partitionStr := "ALTER TABLE %s DROP IF EXISTS PARTITION (`billing_period_start`='%s',`billing_period_end`='%s')"
	stmt := fmt.Sprintf(partitionStr, tableName, start, end)
	err := queryer.Query(stmt)
	return err
}

// hiveName is the identi***REMOVED***er used for Hive columns.
func hiveName(c aws.Column) string {
	name := fmt.Sprintf("%s_%s", strings.TrimSpace(c.Category), strings.TrimSpace(c.Name))
	// hive does not allow ':' or '.' in identi***REMOVED***ers
	name = strings.Replace(name, ":", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return strings.ToLower(name)
}

// hiveType is the data type a column is created as in Hive.
func hiveType(c aws.Column) string {
	switch hiveName(c) {
	case "lineitem_usagestartdate", "lineitem_usageenddate":
		return "timestamp"
	case "lineitem_blendedcost":
		return "double"
	default:
		return "string"
	}
}

package hive

import (
	"fmt"
	"strconv"
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
func CreateAWSUsageTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string, manifests []*aws.Manifest) error {
	location, err := s3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return err
	}

	// Since the billing data likely exists already, we need to enumerate all
	// columns for all manifests to get the entire set of columns used
	// historically.
	// TODO(chance): We will likely want to do this when we add partitions
	// to avoid having to do it all up front.
	columns := make([]Column, 0)
	seen := make(map[string]struct{})
	for _, manifest := range manifests {
		manifestColumns := awsBillingColumns(manifest.Columns)
		for _, col := range manifestColumns {
			if _, exists := seen[col.Name]; !exists {
				seen[col.Name] = struct{}{}
				columns = append(columns, col)
			}
		}
	}

	query := createTable(tableName, location, awsUsageSerde, awsUsageSerdeProps, columns, awsPartitions, true, true)
	return queryer.Query(query)
}

const hiveTimestampLayout = "2006-01-02 15:04:05.000000000"
const hiveDateStringLayout = "20060102"

func UpdateAWSUsageTable(queryer Queryer, tableName, bucket, pre***REMOVED***x string, manifests []*aws.Manifest) error {
	partitionStr := "PARTITION (`billing_period_start`='%s',`billing_period_end`='%s') LOCATION '%s'"
	var stmts []string
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := s3Location(bucket, manifestPath)
		if err != nil {
			return err
		}
		stmt := fmt.Sprintf(partitionStr,
			manifest.BillingPeriod.Start.Format(hiveDateStringLayout),
			manifest.BillingPeriod.End.Format(hiveDateStringLayout),
			location,
		)
		stmts = append(stmts, stmt)

	}
	if len(stmts) == 0 {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE %s DROP IF EXISTS PARTITION (`billing_period_start`!='',`billing_period_end`!='')", tableName)
	err := queryer.Query(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("ALTER TABLE %s ADD %s", tableName, strings.Join(stmts, " "))
	err = queryer.Query(query)
	if err != nil {
		return err
	}

	return nil
}

// hiveName is the identi***REMOVED***er used for Hive columns.
func hiveName(c aws.Column) string {
	name := fmt.Sprintf("%s_%s", c.Category, c.Name)
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

// Columns returns a map of hive column name to it's hive column type.
// Duplicate columns will be suf***REMOVED***xed by an incrementing ordinal. This can
// happen with user de***REMOVED***ned ***REMOVED***elds like tags.
func awsBillingColumns(cols []aws.Column) []Column {
	out := make([]Column, 0)
	seen := make(map[string]int, len(cols))

	for _, c := range cols {
		name := hiveName(c)
		colType := hiveType(c)

		// prevent duplicates by numbering them
		times, exists := seen[name]
		seen[name] = times + 1

		if exists {
			name += strconv.Itoa(times)
		}

		out = append(out, Column{
			Name: name,
			Type: colType,
		})
	}
	return out
}

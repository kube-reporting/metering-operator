package hive

import (
	"fmt"
	"strconv"
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

	awsPartitions = []Column{
		{Name: "billing_period_start", Type: "string"},
		{Name: "billing_period_end", Type: "string"},
	}
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func CreateAWSUsageTable(queryer Queryer, tableName, bucket, prefix string, manifest *aws.Manifest) error {
	location, err := s3Location(bucket, prefix)
	if err != nil {
		return err
	}

	columns := awsBillingColumns(manifest.Columns)

	query := createTable(tableName, location, AWSUsageSerde, AWSUsageSerdeProps, columns, awsPartitions, true, true)
	return queryer.Query(query)
}

const hiveTimestampLayout = "2006-01-02 15:04:05.000000000"
const hiveDateStringLayout = "20060102"

func UpdateAWSUsageTable(queryer Queryer, tableName, bucket, prefix string, manifests []*aws.Manifest) error {
	partitionStr := "PARTITION (`billing_period_start`='%s',`billing_period_end`='%s') LOCATION '%s'"
	var stmts []string
	// A map containing locations we've already added to ensure we do not
	// attempt to add a partition multiple times
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
			stmt := fmt.Sprintf(partitionStr,
				manifest.BillingPeriod.Start.Format(hiveDateStringLayout),
				manifest.BillingPeriod.End.Format(hiveDateStringLayout),
				location,
			)
			stmts = append(stmts, stmt)
		}

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

// hiveName is the identifier used for Hive columns.
func hiveName(c aws.Column) string {
	name := fmt.Sprintf("%s_%s", c.Category, c.Name)
	// hive does not allow ':' or '.' in identifiers
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
// Duplicate columns will be suffixed by an incrementing ordinal. This can
// happen with user defined fields like tags.
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

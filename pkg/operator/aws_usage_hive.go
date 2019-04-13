package operator

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
)

const (
	// AWSUsageHiveSerde is the Hadoop serialization/deserialization implementation used with AWS billing data.
	AWSUsageHiveSerde = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"
)

var (
	// AWSUsageHiveSerdeProps con***REMOVED***gure the SerDe used with AWS Billing Data.
	AWSUsageHiveSerdeProps = map[string]string{
		"serialization.format": ",",
		"***REMOVED***eld.delim":          ",",
		"collection.delim":     "unde***REMOVED***ned",
		"mapkey.delim":         "unde***REMOVED***ned",
		"timestamp.formats":    "yyyy-MM-dd'T'HH:mm:ssZ",
	}

	AWSUsageHivePartitions = []hive.Column{
		{Name: "billing_period_start", Type: "string"},
		{Name: "billing_period_end", Type: "string"},
	}
)

// CreateAWSUsageTable instantiates a new external HiveTable CR for AWS Billing/Usage reports stored in S3.
func (op *Reporting) createAWSUsageHiveTableCR(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName, bucket, pre***REMOVED***x string, manifests []*aws.Manifest) (*cbTypes.HiveTable, error) {
	location, err := hive.S3Location(bucket, pre***REMOVED***x)
	if err != nil {
		return nil, err
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
			name := SanetizeAWSColumnForHive(c)
			colType := AWSColumnToHiveColumnType(c)

			if _, exists := seen[name]; !exists {
				seen[name] = struct{}{}
				columns = append(columns, hive.Column{
					Name: name,
					Type: colType,
				})
			}
		}
	}

	var dbName string
	if dataSource.Spec.AWSBilling.DatabaseName == "" {
		hiveStorage, err := op.getHiveStorage(nil, dataSource.Namespace)
		if err != nil {
			return nil, fmt.Errorf("storage incorrectly con***REMOVED***gured for ReportDataSource %s, err: %s", dataSource.Name, err)
		}
		if hiveStorage.Spec.Hive.DatabaseName == "" {
			return nil, fmt.Errorf("StorageLocation %s Hive database %s does not exist yet", hiveStorage.Name, hiveStorage.Spec.Hive.DatabaseName)
		}
		dbName = hiveStorage.Status.Hive.DatabaseName
	} ***REMOVED*** {
		dbName = dataSource.Spec.AWSBilling.DatabaseName
	}

	if dbName == "" {
		panic(fmt.Sprintf("unable to get dbName for ReportDataSource: %s: should properly return error when database cannot be determined", dataSource.Name))
	}

	params := hive.TableParameters{
		Database:           dbName,
		Name:               tableName,
		Columns:            columns,
		PartitionedBy:      AWSUsageHivePartitions,
		Location:           location,
		FileFormat:         "text***REMOVED***le",
		SerdeFormat:        AWSUsageHiveSerde,
		SerdeRowProperties: AWSUsageHiveSerdeProps,
		External:           true,
	}

	logger.Infof("creating Hive table %s", tableName)
	hiveTable, err := op.createHiveTableCR(dataSource, cbTypes.ReportDataSourceGVK, params, true, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Hive table for ReportDataSource %s: %s", dataSource.Name, err)
	}
	hiveTable, err = op.waitForHiveTable(hiveTable.Namespace, hiveTable.Name, time.Second, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error creating Hive table for ReportDataSource %s: %s", dataSource.Name, err)
	}
	_, err = op.waitForPrestoTable(hiveTable.Namespace, hiveTable.Name, time.Second, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
	}

	logger.Infof("created Hive table %s", tableName)

	return hiveTable, nil
}

func (op *Reporting) updateAWSBillingPartitions(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource, source *cbTypes.S3Bucket, hiveTable *cbTypes.HiveTable, manifests []*aws.Manifest) error {
	logger.Infof("updating partitions for Hive table %s", hiveTable.Name)
	// Fetch the billing manifests
	if len(manifests) == 0 {
		logger.Warnf("HiveTable %q has no report manifests in its bucket, the ***REMOVED***rst report has likely not been generated yet", hiveTable.Name)
		return nil
	}

	var err error
	hiveTable.Spec.Partitions, err = getDesiredPartitions(source.Bucket, manifests)
	if err != nil {
		return err
	}

	_, err = op.meteringClient.MeteringV1alpha1().HiveTables(hiveTable.Namespace).Update(hiveTable)
	if err != nil {
		logger.WithError(err).Errorf("failed to update HiveTable %s partitions for ReportDataSource %s: %s", hiveTable.Name, dataSource.Name, err)
		return err
	}

	return nil
}

func getDesiredPartitions(bucket string, manifests []*aws.Manifest) ([]cbTypes.HiveTablePartition, error) {
	desiredPartitions := make([]cbTypes.HiveTablePartition, 0)
	// Manifests have a one-to-one correlation with hive currentPartitions
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := hive.S3Location(bucket, manifestPath)
		if err != nil {
			return nil, err
		}

		start := reportingutil.AWSBillingPeriodTimestamp(manifest.BillingPeriod.Start.Time)
		end := reportingutil.AWSBillingPeriodTimestamp(manifest.BillingPeriod.End.Time)
		p := cbTypes.HiveTablePartition{
			Location: location,
			PartitionSpec: hive.PartitionSpec{
				"start": start,
				"end":   end,
			},
		}
		desiredPartitions = append(desiredPartitions, p)
	}
	return desiredPartitions, nil
}

// SanetizeAWSColumnForHive removes and replaces invalid characters in AWS
// billing columns with characters allowed in hive SQL
func SanetizeAWSColumnForHive(col aws.Column) string {
	name := fmt.Sprintf("%s_%s", strings.TrimSpace(col.Category), strings.TrimSpace(col.Name))
	// hive does not allow ':' or '.' in identi***REMOVED***ers
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

package operator

import (
	"github.com/sirupsen/logrus"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
)

// CreateAWSUsageTable instantiates a new external Hive table for AWS Billing/Usage reports stored in S3.
func (op *Reporting) createAWSUsageTable(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName, bucket, pre***REMOVED***x string, manifests []*aws.Manifest) error {
	location, err := hive.S3Location(bucket, pre***REMOVED***x)
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
			name := reportingutil.SanetizeAWSColumnForHive(c)
			colType := reportingutil.AWSColumnToHiveColumnType(c)

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
		Partitions:   reportingutil.AWSUsageHivePartitions,
		IgnoreExists: true,
	}
	properties := hive.TableProperties{
		Location:           location,
		FileFormat:         "text***REMOVED***le",
		SerdeFormat:        reportingutil.AWSUsageHiveSerde,
		SerdeRowProperties: reportingutil.AWSUsageHiveSerdeProps,
		External:           true,
	}
	return op.createTableWith(logger, dataSource, cbTypes.SchemeGroupVersion.WithKind("ReportDataSource"), params, properties)
}

package operator

import (
	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/sirupsen/logrus"
)

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
			name := reporting.SanetizeAWSColumnForHive(c)
			colType := reporting.AWSColumnToHiveColumnType(c)

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
		Partitions:   reporting.AWSUsageHivePartitions,
		IgnoreExists: true,
	}
	properties := hive.TableProperties{
		Location:           location,
		FileFormat:         "textfile",
		SerdeFormat:        reporting.AWSUsageHiveSerde,
		SerdeRowProperties: reporting.AWSUsageHiveSerdeProps,
		External:           true,
	}
	return op.createTableWith(logger, dataSource, cbTypes.SchemeGroupVersion.WithKind("ReportDataSource"), params, properties)
}

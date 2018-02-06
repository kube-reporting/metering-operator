package chargeback

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func dataSourceNameToPrestoTableName(name string) string {
	return strings.Replace(dataSourceTableName(name), "_", "-", -1)
}

func (c *Chargeback) runPrestoTableWorker(stopCh <-chan struct{}) {
	logger := c.logger.WithField("component", "prestoTableWorker")
	logger.Infof("PrestoTable worker started")

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-stopCh:
			logger.Infof("PrestoTableWorker exiting")
			return
		case <-ticker.C:
			datasources, err := c.informers.reportDataSourceLister.ReportDataSources(c.cfg.Namespace).List(labels.Everything())
			if err != nil {
				logger.WithError(err).Errorf("unable to list datasources")
				continue
			}
			for _, d := range datasources {
				if d.Spec.AWSBilling != nil {
					err := c.updateAWSBillingPartitions(logger, d)
					if err != nil {
						logger.WithError(err).Errorf("unable to update partitions for datasource %q", d.Name)
					}
				}
			}
		case datasource := <-c.prestoTablePartitionQueue:
			if datasource.Spec.AWSBilling == nil {
				logger.Errorf("incorrectly con***REMOVED***gured datasource sent to the presto table worker: %q", datasource.Name)
				continue
			}
			err := c.updateAWSBillingPartitions(logger, datasource)
			if err != nil {
				logger.WithError(err).Errorf("unable to update partitions for datasource %q", datasource.Name)
			}
		}
	}
}

func (c *Chargeback) updateAWSBillingPartitions(logger log.FieldLogger, datasource *cbTypes.ReportDataSource) error {
	prestoTableName := dataSourceNameToPrestoTableName(datasource.Name)
	prestoTable, err := c.informers.prestoTableLister.PrestoTables(c.cfg.Namespace).Get(prestoTableName)
	// If this came over the work queue, the presto table may not be in the
	// cache, so check if it exists via the API before erroring out
	if k8serrors.IsNotFound(err) {
		getOpts := meta.GetOptions{}
		prestoTable, err = c.chargebackClient.ChargebackV1alpha1().PrestoTables(c.cfg.Namespace).Get(prestoTableName, getOpts)
	}
	if err != nil {
		return err
	}
	prestoTable = prestoTable.DeepCopy()

	logger.Infof("updating partitions for presto table %s", prestoTable.Name)

	// Fetch the billing manifests
	manifestRetriever, err := aws.NewManifestRetriever(datasource.Spec.AWSBilling.Source.Bucket, datasource.Spec.AWSBilling.Source.Pre***REMOVED***x)
	if err != nil {
		return err
	}

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("prestoTable %q has no report manifests in its bucket, the ***REMOVED***rst report has likely not been generated yet", prestoTable.Name)
		return nil
	}

	// Compare the manifests list and existing partitions, deleting stale
	// partitions and creating missing partitions
	currentPartitions := prestoTable.State.Partitions
	desiredPartitions, err := getDesiredPartitions(prestoTable, datasource, manifests)
	if err != nil {
		return err
	}

	changes := getPartitionChanges(currentPartitions, desiredPartitions)

	currentPartitionsList := make([]string, len(currentPartitions))
	desiredPartitionsList := make([]string, len(desiredPartitions))
	toRemovePartitionsList := make([]string, len(changes.toRemovePartitions))
	toAddPartitionsList := make([]string, len(changes.toAddPartitions))

	for i, p := range currentPartitions {
		currentPartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range desiredPartitions {
		desiredPartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range changes.toRemovePartitions {
		toRemovePartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range changes.toAddPartitions {
		toAddPartitionsList[i] = fmt.Sprintf("%#v", p)
	}

	logger.Debugf("current partitions: %s", strings.Join(currentPartitionsList, ", "))
	logger.Debugf("desired partitions: %s", strings.Join(desiredPartitionsList, ", "))
	logger.Debugf("partitions to remove: %s", strings.Join(toRemovePartitionsList, ", "))
	logger.Debugf("partitions to add: %s", strings.Join(toAddPartitionsList, ", "))

	// We do removals then additions so that updates are supported as a combination of remove + add partition
	for _, p := range changes.toRemovePartitions {
		logger.Warnf("Deleting partition from presto table %q with range %s-%s, this is unexpected", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
		err = hive.DropPartition(c.hiveQueryer, prestoTable.State.CreationParameters.TableName, p.Start, p.End)
		if err != nil {
			logger.WithError(err).Errorf("failed to drop partition in table %s for range %s-%s at location %s", prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
			return err
		}
		logger.Debugf("partition successfully deleted from presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
	}

	for _, p := range changes.toAddPartitions {
		// This partition doesn't exist in hive. Create it.
		logger.Debugf("Adding partition to presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
		err = hive.AddPartition(c.hiveQueryer, prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
		if err != nil {
			logger.WithError(err).Errorf("failed to add partition in table %s for range %s-%s at location %s", prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
			return err
		}
		logger.Debugf("partition successfully added to presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
	}

	prestoTable.State.Partitions = desiredPartitions

	_, err = c.chargebackClient.ChargebackV1alpha1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("failed to update PrestoTable CR partitions for %q", prestoTable.Name)
		return err
	}

	logger.Infof("***REMOVED***nished updating partitions for prestoTable %q", prestoTable.Name)

	return nil
}

func getDesiredPartitions(prestoTable *cbTypes.PrestoTable, datasource *cbTypes.ReportDataSource, manifests []*aws.Manifest) ([]cbTypes.PrestoTablePartition, error) {
	desiredPartitions := make([]cbTypes.PrestoTablePartition, 0)
	// Manifests have a one-to-one correlation with hive currentPartitions
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := hive.S3Location(datasource.Spec.AWSBilling.Source.Bucket, manifestPath)
		if err != nil {
			return nil, err
		}
		p := cbTypes.PrestoTablePartition{
			Start:    manifest.BillingPeriod.Start.Format(hive.HiveDateStringLayout),
			End:      manifest.BillingPeriod.End.Format(hive.HiveDateStringLayout),
			Location: location,
		}
		desiredPartitions = append(desiredPartitions, p)
	}
	return desiredPartitions, nil
}

type partitionChanges struct {
	toRemovePartitions []cbTypes.PrestoTablePartition
	toAddPartitions    []cbTypes.PrestoTablePartition
}

func getPartitionChanges(currentPartitions, desiredPartitions []cbTypes.PrestoTablePartition) partitionChanges {
	currentPartitionsSet := make(map[string]cbTypes.PrestoTablePartition)
	desiredPartitionsSet := make(map[string]cbTypes.PrestoTablePartition)

	for _, p := range currentPartitions {
		currentPartitionsSet[fmt.Sprintf("%s_%s", p.Start, p.End)] = p
	}
	for _, p := range desiredPartitions {
		desiredPartitionsSet[fmt.Sprintf("%s_%s", p.Start, p.End)] = p
	}

	var toRemovePartitions, toAddPartitions []cbTypes.PrestoTablePartition

	for key, partition := range currentPartitionsSet {
		if newPartition, exists := desiredPartitionsSet[key]; !exists || (newPartition.Location != partition.Location) {
			toRemovePartitions = append(toRemovePartitions, partition)
		}
	}
	for key, partition := range desiredPartitionsSet {
		if newPartition, exists := currentPartitionsSet[key]; !exists || (newPartition.Location != partition.Location) {
			toAddPartitions = append(toAddPartitions, partition)
		}
	}

	return partitionChanges{
		toRemovePartitions: toRemovePartitions,
		toAddPartitions:    toAddPartitions,
	}
}

package chargeback

import (
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
			datasources, err := c.informers.reportDataSourceLister.ReportDataSources(c.namespace).List(labels.Everything())
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
	prestoTable, err := c.informers.prestoTableLister.PrestoTables(c.namespace).Get(prestoTableName)
	// If this came over the work queue, the presto table may not be in the
	// cache, so check if it exists via the API before erroring out
	if k8serrors.IsNotFound(err) {
		getOpts := meta.GetOptions{}
		prestoTable, err = c.chargebackClient.ChargebackV1alpha1().PrestoTables(c.namespace).Get(prestoTableName, getOpts)
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
	partitions := make(map[cbTypes.PrestoTablePartition]struct{})
	//TODO: look into using the k8s.io/apimachinery/util/sets with the
	//k8s.io/code-generator's set-gen generator? Sets would give us typical set
	//operations which might make it easier to read in the future as we can just
	//compute the sets up front
	// https://github.com/coreos-inc/kube-chargeback/pull/141#discussion_r157099880

	for _, p := range prestoTable.State.Partitions {
		partitions[p] = struct{}{}
	}

	// Manifests have a one-to-one correlation with hive partitions
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := hive.S3Location(datasource.Spec.AWSBilling.Source.Bucket, manifestPath)
		if err != nil {
			return err
		}
		p := cbTypes.PrestoTablePartition{
			Start:    manifest.BillingPeriod.Start.Format(hive.HiveDateStringLayout),
			End:      manifest.BillingPeriod.End.Format(hive.HiveDateStringLayout),
			Location: location,
		}
		if _, ok := partitions[p]; ok {
			// This partition exists in hive. Remove it from the map and proceed
			// on to the next manifest.
			delete(partitions, p)
			continue
		}
		// This partition doesn't exist in hive. Create it.
		logger.Debugf("Adding partition to presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
		err = hive.AddPartition(c.hiveQueryer, prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
		if err != nil {
			logger.WithError(err).Errorf("failed to add partition in table %s for range %s-%s at location %s", prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
			return err
		}
		// Update the CR with the new partition, so we remember it next time.
		prestoTable.State.Partitions = append(prestoTable.State.Partitions, p)
		logger.Debugf("partition successfully added to presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
	}

	// Any remaining entries in the map are existing partitions which we didn't
	// ***REMOVED***nd in the manifests, and are thus stale. Delete them.
	for p, _ := range partitions {
		logger.Warnf("Deleting partition from presto table %q with range %s-%s, this is unexpected", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
		err = hive.DropPartition(c.hiveQueryer, prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
		if err != nil {
			logger.WithError(err).Errorf("failed to drop partition in table %s for range %s-%s at location %s", prestoTable.State.CreationParameters.TableName, p.Start, p.End, p.Location)
			return err
		}

		// Remove the deleted partition from the CT
		removePartition(p, prestoTable)
		logger.Debugf("partition successfully deleted from presto table %q with range %s-%s", prestoTable.State.CreationParameters.TableName, p.Start, p.End)
	}

	_, err = c.chargebackClient.ChargebackV1alpha1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("failed to update PrestoTable CR partitions for %q", prestoTable.Name)
		return err
	}

	logger.Infof("***REMOVED***nished updating partitions for prestoTable %q", prestoTable.Name)

	return nil
}

func removePartition(p cbTypes.PrestoTablePartition, table *cbTypes.PrestoTable) {
	for i := len(table.State.Partitions) - 1; i >= 0; i-- {
		if table.State.Partitions[i] == p {
			table.State.Partitions = append(
				table.State.Partitions[:i],
				table.State.Partitions[i+1:]...)
		}
	}
}

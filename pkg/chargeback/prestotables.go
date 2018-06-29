package chargeback

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

const prestoTableReconcileInterval = time.Minute

func dataSourceNameToPrestoTableName(name string) string {
	return strings.Replace(dataSourceTableName(name), "_", "-", -1)
}

func (c *Chargeback) runPrestoTableWorker(stopCh <-chan struct{}) {
	logger := c.logger.WithField("component", "prestoTableWorker")
	logger.Infof("PrestoTable worker started")

	for {
		select {
		case <-stopCh:
			logger.Infof("PrestoTableWorker exiting")
			return
		case <-c.clock.Tick(prestoTableReconcileInterval):
			datasources, err := c.informers.Chargeback().V1alpha1().ReportDataSources().Lister().ReportDataSources(c.cfg.Namespace).List(labels.Everything())
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
	prestoTableResourceName := prestoTableResourceNameFromKind("datasource", datasource.Name)
	prestoTable, err := c.informers.Chargeback().V1alpha1().PrestoTables().Lister().PrestoTables(c.cfg.Namespace).Get(prestoTableResourceName)
	// If this came over the work queue, the presto table may not be in the
	// cache, so check if it exists via the API before erroring out
	if k8serrors.IsNotFound(err) {
		prestoTable, err = c.chargebackClient.ChargebackV1alpha1().PrestoTables(c.cfg.Namespace).Get(prestoTableResourceName, metav1.GetOptions{})
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
	toUpdatePartitionsList := make([]string, len(changes.toUpdatePartitions))

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
	for i, p := range changes.toUpdatePartitions {
		toUpdatePartitionsList[i] = fmt.Sprintf("%#v", p)
	}

	logger.Debugf("current partitions: %s", strings.Join(currentPartitionsList, ", "))
	logger.Debugf("desired partitions: %s", strings.Join(desiredPartitionsList, ", "))
	logger.Debugf("partitions to remove: [%s]", strings.Join(toRemovePartitionsList, ", "))
	logger.Debugf("partitions to add: [%s]", strings.Join(toAddPartitionsList, ", "))
	logger.Debugf("partitions to update: [%s]", strings.Join(toUpdatePartitionsList, ", "))

	var toRemove []cbTypes.TablePartition = append(changes.toRemovePartitions, changes.toUpdatePartitions...)
	var toAdd []cbTypes.TablePartition = append(changes.toAddPartitions, changes.toUpdatePartitions...)
	// We do removals then additions so that updates are supported as a combination of remove + add partition

	tableName := prestoTable.State.Parameters.Name
	for _, p := range toRemove {
		start := p.PartitionSpec["start"]
		end := p.PartitionSpec["end"]
		logger.Warnf("Deleting partition from presto table %q with range %s-%s", tableName, start, end)
		err = dropAWSHivePartition(c.hiveQueryer, tableName, start, end)
		if err != nil {
			logger.WithError(err).Errorf("failed to drop partition in table %s for range %s-%s", tableName, start, end)
			return err
		}
		logger.Debugf("partition successfully deleted from presto table %q with range %s-%s", tableName, start, end)
	}

	for _, p := range toAdd {
		start := p.PartitionSpec["start"]
		end := p.PartitionSpec["end"]
		// This partition doesn't exist in hive. Create it.
		logger.Debugf("Adding partition to presto table %q with range %s-%s", tableName, start, end)
		err = addAWSHivePartition(c.hiveQueryer, tableName, start, end, p.Location)
		if err != nil {
			logger.WithError(err).Errorf("failed to add partition in table %s for range %s-%s at location %s", prestoTable.State.Parameters.Name, p.PartitionSpec["start"], p.PartitionSpec["end"], p.Location)
			return err
		}
		logger.Debugf("partition successfully added to presto table %q with range %s-%s", tableName, start, end)
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

func getDesiredPartitions(prestoTable *cbTypes.PrestoTable, datasource *cbTypes.ReportDataSource, manifests []*aws.Manifest) ([]cbTypes.TablePartition, error) {
	desiredPartitions := make([]cbTypes.TablePartition, 0)
	// Manifests have a one-to-one correlation with hive currentPartitions
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := hive.S3Location(datasource.Spec.AWSBilling.Source.Bucket, manifestPath)
		if err != nil {
			return nil, err
		}

		start := billingPeriodTimestamp(manifest.BillingPeriod.Start.Time)
		end := billingPeriodTimestamp(manifest.BillingPeriod.End.Time)
		p := cbTypes.TablePartition{
			Location: location,
			PartitionSpec: presto.PartitionSpec{
				"start": start,
				"end":   end,
			},
		}
		desiredPartitions = append(desiredPartitions, p)
	}
	return desiredPartitions, nil
}

type partitionChanges struct {
	toRemovePartitions []cbTypes.TablePartition
	toAddPartitions    []cbTypes.TablePartition
	toUpdatePartitions []cbTypes.TablePartition
}

func getPartitionChanges(currentPartitions, desiredPartitions []cbTypes.TablePartition) partitionChanges {
	currentPartitionsSet := make(map[string]cbTypes.TablePartition)
	desiredPartitionsSet := make(map[string]cbTypes.TablePartition)

	for _, p := range currentPartitions {
		currentPartitionsSet[fmt.Sprintf("%s_%s", p.PartitionSpec["start"], p.PartitionSpec["end"])] = p
	}
	for _, p := range desiredPartitions {
		desiredPartitionsSet[fmt.Sprintf("%s_%s", p.PartitionSpec["start"], p.PartitionSpec["end"])] = p
	}

	var toRemovePartitions, toAddPartitions, toUpdatePartitions []cbTypes.TablePartition

	for key, partition := range currentPartitionsSet {
		if _, exists := desiredPartitionsSet[key]; !exists {
			toRemovePartitions = append(toRemovePartitions, partition)
		}
	}
	for key, partition := range desiredPartitionsSet {
		if _, exists := currentPartitionsSet[key]; !exists {
			toAddPartitions = append(toAddPartitions, partition)
		}
	}
	for key, existingPartition := range currentPartitionsSet {
		if newPartition, exists := desiredPartitionsSet[key]; exists && (newPartition.Location != existingPartition.Location) {
			// use newPartition so toUpdatePartitions contains the desired partition state
			toUpdatePartitions = append(toUpdatePartitions, newPartition)
		}
	}

	return partitionChanges{
		toRemovePartitions: toRemovePartitions,
		toAddPartitions:    toAddPartitions,
		toUpdatePartitions: toUpdatePartitions,
	}
}

func (c *Chargeback) createPrestoTableCR(obj runtime.Object, apiVersion, kind string, params hive.TableParameters, properties hive.TableProperties, partitions []presto.TablePartition) error {
	accessor := meta.NewAccessor()
	name, err := accessor.Name(obj)
	if err != nil {
		return err
	}
	uid, err := accessor.UID(obj)
	if err != nil {
		return err
	}
	namespace, err := accessor.Namespace(obj)
	if err != nil {
		return err
	}
	objLabels, err := accessor.Labels(obj)
	if err != nil {
		return err
	}

	resourceName := prestoTableResourceNameFromKind(kind, name)
	prestoTableCR := cbTypes.PrestoTable{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PrestoTable",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels:    objLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiVersion,
					Kind:       kind,
					Name:       name,
					UID:        uid,
				},
			},
		},
		State: cbTypes.PrestoTableState{
			Parameters: cbTypes.TableParameters(hive.TableParameters{
				Name:         params.Name,
				Columns:      params.Columns,
				IgnoreExists: params.IgnoreExists,
				Partitions:   params.Partitions,
			}),
			Properties: cbTypes.TableProperties(hive.TableProperties{
				Location:           properties.Location,
				FileFormat:         properties.FileFormat,
				SerdeFormat:        properties.SerdeFormat,
				SerdeRowProperties: properties.SerdeRowProperties,
				External:           properties.External,
			}),
		},
	}
	for _, partition := range partitions {
		prestoTableCR.State.Partitions = append(prestoTableCR.State.Partitions, cbTypes.TablePartition(partition))
	}

	client := c.chargebackClient.ChargebackV1alpha1().PrestoTables(namespace)
	_, err = client.Create(&prestoTableCR)
	if k8serrors.IsAlreadyExists(err) {
		if existing, err := client.Get(resourceName, metav1.GetOptions{}); err != nil {
			return err
		} ***REMOVED*** {
			prestoTableCR.ResourceVersion = existing.ResourceVersion
			_, err = client.Update(&prestoTableCR)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

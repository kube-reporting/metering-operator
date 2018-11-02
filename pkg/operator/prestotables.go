package operator

import (
	"strings"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	prestoTableFinalizer = cbTypes.GroupName + "/prestotable"
)

func dataSourceNameToPrestoTableName(name string) string {
	return strings.Replace(reportingutil.DataSourceTableName(name), "_", "-", -1)
}

func (op *Reporting) runPrestoTableWorker(stopCh <-chan struct{}) {
	logger := op.logger.WithField("component", "prestoTableWorker")
	logger.Infof("PrestoTable worker started")

	for op.processPrestoTable(logger) {

	}

}
func (op *Reporting) processPrestoTable(logger log.FieldLogger) bool {
	obj, quit := op.prestoTableQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting PrestoTable worker")
		return false
	}
	defer op.prestoTableQueue.Done(obj)

	logger = logger.WithFields(newLogIdentifier(op.rand))
	if key, ok := op.getKeyFromQueueObj(logger, "PrestoTable", obj, op.prestoTableQueue); ok {
		err := op.syncPrestoTable(logger, key)
		const maxRequeues = 10
		op.handleErr(logger, err, "PrestoTable", key, op.prestoTableQueue, maxRequeues)
	}
	return true
}

func (op *Reporting) syncPrestoTable(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("PrestoTable", name)

	prestoTableLister := op.prestoTableLister
	prestoTable, err := prestoTableLister.PrestoTables(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("PrestoTable %s does not exist anymore", key)
			return nil
		}
		return err
	}
	pt := prestoTable.DeepCopy()

	if pt.DeletionTimestamp != nil {
		logger.Infof("PrestoTable is marked for deletion, performing cleanup")
		err := op.dropPrestoTable(pt)
		if err != nil {
			return err
		}
		_, err = op.removePrestoTableFinalizer(pt)
		return err
	}

	logger.Infof("syncing PrestoTable %s", pt.GetName())
	err = op.handlePrestoTable(logger, pt)
	if err != nil {
		logger.WithError(err).Errorf("error syncing PrestoTable %s", pt.GetName())
		return err
	}
	logger.Infof("successfully synced PrestoTable %s", pt.GetName())
	return nil
}

func (op *Reporting) handlePrestoTable(logger log.FieldLogger, prestoTable *cbTypes.PrestoTable) error {
	if op.cfg.EnableFinalizers && prestoTableNeedsFinalizer(prestoTable) {
		var err error
		prestoTable, err = op.addPrestoTableFinalizer(prestoTable)
		if err != nil {
			return err
		}
	}

	return nil
}

func (op *Reporting) createPrestoTableCR(obj metav1.Object, gvk schema.GroupVersionKind, params hive.TableParameters, properties hive.TableProperties, partitions []presto.TablePartition) error {
	apiVersion := gvk.GroupVersion().String()
	kind := gvk.Kind
	name := obj.GetName()
	namespace := obj.GetNamespace()
	objLabels := obj.GetLabels()
	ownerRef := metav1.NewControllerRef(obj, gvk)

	var finalizers []string
	if op.cfg.EnableFinalizers {
		finalizers = []string{prestoTableFinalizer}
	}

	resourceName := reportingutil.PrestoTableResourceNameFromKind(kind, name)
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
				*ownerRef,
			},
			Finalizers: finalizers,
		},
		Status: cbTypes.PrestoTableStatus{
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
		prestoTableCR.Status.Partitions = append(prestoTableCR.Status.Partitions, cbTypes.TablePartition(partition))
	}

	_, err := op.meteringClient.MeteringV1alpha1().PrestoTables(namespace).Create(&prestoTableCR)
	return err
}

func (op *Reporting) addPrestoTableFinalizer(prestoTable *cbTypes.PrestoTable) (*cbTypes.PrestoTable, error) {
	prestoTable.Finalizers = append(prestoTable.Finalizers, prestoTableFinalizer)
	newPrestoTable, err := op.meteringClient.MeteringV1alpha1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	logger := op.logger.WithField("PrestoTable", prestoTable.Name)
	if err != nil {
		logger.WithError(err).Errorf("error adding %s finalizer to PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
		return nil, err
	}
	logger.Infof("added %s finalizer to PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
	return newPrestoTable, nil
}

func (op *Reporting) removePrestoTableFinalizer(prestoTable *cbTypes.PrestoTable) (*cbTypes.PrestoTable, error) {
	if !slice.ContainsString(prestoTable.ObjectMeta.Finalizers, prestoTableFinalizer, nil) {
		return prestoTable, nil
	}
	prestoTable.Finalizers = slice.RemoveString(prestoTable.Finalizers, prestoTableFinalizer, nil)
	newPrestoTable, err := op.meteringClient.MeteringV1alpha1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	logger := op.logger.WithField("prestoTable", prestoTable.Name)
	if err != nil {
		logger.WithError(err).Errorf("error removing %s finalizer from PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
		return nil, err
	}
	logger.Infof("removed %s finalizer from PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
	return newPrestoTable, nil
}

func prestoTableNeedsFinalizer(prestoTable *cbTypes.PrestoTable) bool {
	return prestoTable.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(prestoTable.ObjectMeta.Finalizers, prestoTableFinalizer, nil)
}

func (op *Reporting) dropPrestoTable(prestoTable *cbTypes.PrestoTable) error {
	tableName := prestoTable.Status.Parameters.Name
	logger := op.logger.WithFields(log.Fields{"PrestoTable": prestoTable.Name, "tableName": tableName})
	logger.Infof("dropping presto table %s", tableName)
	err := op.tableManager.DropTable(tableName, true)
	if err != nil {
		logger.WithError(err).Error("unable to drop presto table")
		return err
	}
	logger.Infof("successfully deleted table %s", tableName)
	return nil
}

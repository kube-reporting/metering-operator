package operator

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	prestoTableFinalizer = metering.GroupName + "/prestotable"
)

func (op *Reporting) runPrestoTableWorker() {
	logger := op.logger.WithField("component", "prestoTableWorker")
	logger.Infof("PrestoTable worker started")
	const maxRequeues = 10
	for op.processResource(logger, op.syncPrestoTable, "PrestoTable", op.prestoTableQueue, maxRequeues) {
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

	logger = logger.WithFields(log.Fields{"prestoTable": name, "namespace": namespace})

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
		if !pt.Spec.Unmanaged {
			err := op.dropPrestoTable(pt)
			if err != nil {
				return err
			}
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

func (op *Reporting) handlePrestoTable(logger log.FieldLogger, prestoTable *metering.PrestoTable) error {
	if op.cfg.EnableFinalizers && prestoTableNeedsFinalizer(prestoTable) {
		var err error
		prestoTable, err = op.addPrestoTableFinalizer(prestoTable)
		if err != nil {
			return err
		}
	}

	if prestoTable.Spec.Catalog == "" {
		return fmt.Errorf("spec.catalog must be set")
	}
	if prestoTable.Spec.Schema == "" {
		return fmt.Errorf("spec.schema must be set")
	}
	if prestoTable.Spec.TableName == "" {
		return fmt.Errorf("spec.tableName must be set")
	}
	if !prestoTable.Spec.Unmanaged && len(prestoTable.Spec.Columns) == 0 {
		return fmt.Errorf("spec.columns must be non-empty when spec.unmanaged is set to false")
	}
	if prestoTable.Spec.CreateTableAs && prestoTable.Spec.View {
		return fmt.Errorf("spec.createTableAs and spec.view are mutually exclusive")
	}

	if prestoTable.Status.TableName != "" {
		logger.Infof("PrestoTable %s status.tableName already set to %s, skipping", prestoTable.Name, prestoTable.Status.TableName)
		return nil
	}

	var (
		needsUpdate bool
		err         error
	)
	if prestoTable.Spec.Unmanaged {
		logger.Infof("PrestoTable %s is unmanaged", prestoTable.Name)

		prestoTable.Spec.Columns, err = op.prestoTableManager.QueryMetadata(prestoTable.Spec.Catalog, prestoTable.Spec.Schema, prestoTable.Spec.TableName)
		if err != nil {
			return fmt.Errorf("failed to query the %s Presto table metadata: %v", prestoTable.Spec.TableName, err)
		}

		needsUpdate = copyPrestoTableSpecToStatus(prestoTable)
	} else {
		var tableStr string
		switch {
		case prestoTable.Spec.View:
			tableStr = "view"
			logger.Infof("creating view %s", prestoTable.Spec.TableName)
			err = op.prestoTableManager.CreateView(prestoTable.Spec.Catalog, prestoTable.Spec.Schema, prestoTable.Spec.TableName, prestoTable.Spec.Query)
		case !prestoTable.Spec.CreateTableAs:
			logger.Infof("creating table %s", prestoTable.Spec.TableName)
			tableStr = "table"
			err = op.prestoTableManager.CreateTable(prestoTable.Spec.Catalog, prestoTable.Spec.Schema, prestoTable.Spec.TableName, prestoTable.Spec.Columns, prestoTable.Spec.Comment, prestoTable.Spec.Properties, true)
		case prestoTable.Spec.CreateTableAs:
			logger.Infof("creating table %s", prestoTable.Spec.TableName)
			tableStr = "table"
			err = op.prestoTableManager.CreateTableAs(prestoTable.Spec.Catalog, prestoTable.Spec.Schema, prestoTable.Spec.TableName, prestoTable.Spec.Columns, prestoTable.Spec.Comment, prestoTable.Spec.Properties, false, prestoTable.Spec.Query)
		default:
			panic("unhandled case for PrestoTable")
		}
		if err != nil {
			return fmt.Errorf("error creating %s %s: %s", tableStr, prestoTable.Spec.TableName, err)
		}
		logger.Infof("created %s %s", tableStr, prestoTable.Spec.TableName)
		needsUpdate = copyPrestoTableSpecToStatus(prestoTable)
	}

	if needsUpdate {
		var err error
		prestoTable, err = op.meteringClient.MeteringV1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
		if err != nil {
			return fmt.Errorf("unable to update PrestoTable %s status: %s", prestoTable.Name, err)
		}
		if err := op.queueDependentsOfPrestoTable(prestoTable); err != nil {
			logger.WithError(err).Errorf("error queuing dependents of PrestoTable %s", prestoTable.Name)
		}
	}
	return nil
}

func copyPrestoTableSpecToStatus(prestoTable *metering.PrestoTable) bool {
	var needsUpdate bool
	if prestoTable.Status.Catalog != prestoTable.Spec.Catalog {
		prestoTable.Status.Catalog = prestoTable.Spec.Catalog
		needsUpdate = true
	}
	if prestoTable.Status.Schema != prestoTable.Spec.Schema {
		prestoTable.Status.Schema = prestoTable.Spec.Schema
		needsUpdate = true
	}
	if prestoTable.Status.TableName != prestoTable.Spec.TableName {
		prestoTable.Status.TableName = prestoTable.Spec.TableName
		needsUpdate = true
	}
	// Create copies so we do not mutate the original slices
	specColCopy := make([]presto.Column, len(prestoTable.Spec.Columns))
	copy(specColCopy, prestoTable.Spec.Columns)
	sort.Slice(specColCopy, func(i, j int) bool {
		return specColCopy[i].Name < specColCopy[j].Name
	})
	statusColCopy := make([]presto.Column, len(prestoTable.Status.Columns))
	copy(statusColCopy, prestoTable.Spec.Columns)
	sort.Slice(statusColCopy, func(i, j int) bool {
		return statusColCopy[i].Name < statusColCopy[j].Name
	})
	if !reflect.DeepEqual(statusColCopy, specColCopy) {
		prestoTable.Status.Columns = prestoTable.Spec.Columns
		needsUpdate = true
	}
	if !reflect.DeepEqual(prestoTable.Status.Properties, prestoTable.Spec.Properties) {
		prestoTable.Status.Properties = prestoTable.Spec.Properties
		needsUpdate = true
	}
	if prestoTable.Status.Comment != prestoTable.Spec.Comment {
		prestoTable.Status.Comment = prestoTable.Spec.Comment
		needsUpdate = true
	}
	if prestoTable.Status.View != prestoTable.Spec.View {
		prestoTable.Status.View = prestoTable.Spec.View
		needsUpdate = true
	}
	if prestoTable.Status.CreateTableAs != prestoTable.Spec.CreateTableAs {
		prestoTable.Status.CreateTableAs = prestoTable.Spec.CreateTableAs
		needsUpdate = true
	}
	if prestoTable.Status.Query != prestoTable.Spec.Query {
		prestoTable.Status.Query = prestoTable.Spec.Query
		needsUpdate = true
	}
	return needsUpdate
}

func (op *Reporting) addPrestoTableFinalizer(prestoTable *metering.PrestoTable) (*metering.PrestoTable, error) {
	prestoTable.Finalizers = append(prestoTable.Finalizers, prestoTableFinalizer)
	newPrestoTable, err := op.meteringClient.MeteringV1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	logger := op.logger.WithFields(log.Fields{"prestoTable": prestoTable.Name, "namespace": prestoTable.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s finalizer to PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
		return nil, err
	}
	logger.Infof("added %s finalizer to PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
	return newPrestoTable, nil
}

func (op *Reporting) removePrestoTableFinalizer(prestoTable *metering.PrestoTable) (*metering.PrestoTable, error) {
	if !slice.ContainsString(prestoTable.ObjectMeta.Finalizers, prestoTableFinalizer, nil) {
		return prestoTable, nil
	}
	prestoTable.Finalizers = slice.RemoveString(prestoTable.Finalizers, prestoTableFinalizer, nil)
	newPrestoTable, err := op.meteringClient.MeteringV1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	logger := op.logger.WithFields(log.Fields{"prestoTable": prestoTable.Name, "namespace": prestoTable.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error removing %s finalizer from PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
		return nil, err
	}
	logger.Infof("removed %s finalizer from PrestoTable: %s/%s", prestoTableFinalizer, prestoTable.Namespace, prestoTable.Name)
	return newPrestoTable, nil
}

func prestoTableNeedsFinalizer(prestoTable *metering.PrestoTable) bool {
	return prestoTable.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(prestoTable.ObjectMeta.Finalizers, prestoTableFinalizer, nil)
}

func (op *Reporting) dropPrestoTable(prestoTable *metering.PrestoTable) error {
	if !prestoTable.Spec.Unmanaged {
		if prestoTable.Status.TableName == "" {
			return nil
		}
		if prestoTable.Status.View {
			return op.prestoTableManager.DropView(prestoTable.Status.Catalog, prestoTable.Status.Schema, prestoTable.Status.TableName, true)
		} else {
			return op.prestoTableManager.DropTable(prestoTable.Status.Catalog, prestoTable.Status.Schema, prestoTable.Status.TableName, true)
		}
	}
	if prestoTable.Spec.Unmanaged {
		return errors.New("cannot drop unmanaged PrestoTable")
	}
	return errors.New("dropping PrestoTables is currently unsupported")
}

func (op *Reporting) createPrestoTableCR(obj metav1.Object, gvk schema.GroupVersionKind, catalog, schema, tableName string, columns []presto.Column, unmanaged, view bool, query string) (*metering.PrestoTable, error) {
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

	resourceName := reportingutil.TableResourceNameFromKind(kind, namespace, name)
	newPrestoTable := &metering.PrestoTable{
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
		Spec: metering.PrestoTableSpec{
			Unmanaged: unmanaged,
			Catalog:   catalog,
			Schema:    schema,
			TableName: tableName,
			Columns:   columns,
			View:      view,
			Query:     query,
		},
	}
	var err error
	prestoTable, err := op.meteringClient.MeteringV1().PrestoTables(namespace).Create(newPrestoTable)
	switch {
	case apierrors.IsAlreadyExists(err):
		op.logger.Warnf("PrestoTable %s already exists", resourceName)
		prestoTable, err = op.prestoTableLister.PrestoTables(namespace).Get(resourceName)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, fmt.Errorf("couldn't create PrestoTable resource for %s %s: %v", gvk, name, err)
	}
	return prestoTable, nil
}

func (op *Reporting) waitForPrestoTable(namespace, name string, pollInterval, timeout time.Duration) (*metering.PrestoTable, error) {
	var prestoTable *metering.PrestoTable
	err := wait.Poll(pollInterval, timeout, func() (bool, error) {
		var err error
		prestoTable, err = op.meteringClient.MeteringV1().PrestoTables(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if prestoTable.Status.TableName != "" {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return nil, errors.New("timed out waiting for Hive table to be created")
		}
		return nil, err
	}
	return prestoTable, nil
}

func (op *Reporting) queueDependentsOfPrestoTable(prestoTable *metering.PrestoTable) error {
	reports, err := op.reportLister.Reports(prestoTable.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	datasources, err := op.reportDataSourceLister.ReportDataSources(prestoTable.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, datasource := range datasources {
		if datasource.Spec.PrestoTable != nil && datasource.Spec.PrestoTable.TableRef.Name == prestoTable.Name {
			op.enqueueReportDataSource(datasource)
		}
	}
	for _, report := range reports {
		if report.Status.TableRef.Name == prestoTable.Name {
			op.enqueueReport(report)
		}
	}
	return nil
}

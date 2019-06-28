package operator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	hiveTableFinalizer = metering.GroupName + "/hivetable"
)

var (
	hiveTablePartitionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metering",
			Name:      "table_partitions",
			Help:      "Current number of partitions in a HiveTable.",
		},
		[]string{"table_name"},
	)
)

func init() {
	prometheus.MustRegister(hiveTablePartitionsGauge)
}

func (op *Reporting) runHiveTableWorker() {
	logger := op.logger.WithField("component", "hiveTableWorker")
	logger.Infof("HiveTable worker started")
	const maxRequeues = 10
	for op.processResource(logger, op.syncHiveTable, "HiveTable", op.hiveTableQueue, maxRequeues) {
	}
}

func (op *Reporting) syncHiveTable(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"hiveTable": name, "namespace": namespace})

	hiveTableLister := op.hiveTableLister
	hiveTable, err := hiveTableLister.HiveTables(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("HiveTable %s does not exist anymore", key)
			return nil
		}
		return err
	}
	pt := hiveTable.DeepCopy()

	if pt.DeletionTimestamp != nil {
		logger.Infof("HiveTable is marked for deletion, performing cleanup")
		err := op.dropHiveTable(pt)
		if err != nil {
			return err
		}
		_, err = op.removeHiveTableFinalizer(pt)
		return err
	}

	logger.Infof("syncing HiveTable %s", pt.GetName())
	err = op.handleHiveTable(logger, pt)
	if err != nil {
		logger.WithError(err).Errorf("error syncing HiveTable %s", pt.GetName())
		return err
	}
	logger.Infof("successfully synced HiveTable %s", pt.GetName())
	return nil
}

func (op *Reporting) handleHiveTable(logger log.FieldLogger, hiveTable *metering.HiveTable) error {
	if op.cfg.EnableFinalizers && hiveTableNeedsFinalizer(hiveTable) {
		var err error
		hiveTable, err = op.addHiveTableFinalizer(hiveTable)
		if err != nil {
			return err
		}
	}
	if hiveTable.Spec.TableName == "" {
		return fmt.Errorf("spec.tableName must be set")
	}
	if len(hiveTable.Spec.Columns) == 0 {
		return fmt.Errorf("spec.columns must be non-empty")
	}

	if hiveTable.Status.TableName != "" {
		logger.Infof("HiveTable %s already created", hiveTable.Name)
	} ***REMOVED*** {
		logger.Infof("creating table %s in Hive", hiveTable.Spec.TableName)

		var newSortedBy []hive.SortColumn
		for _, c := range hiveTable.Spec.SortedBy {
			newSortedBy = append(newSortedBy, hive.SortColumn{
				Name:      c.Name,
				Decending: c.Decending,
			})
		}
		params := hive.TableParameters{
			Database:        hiveTable.Spec.DatabaseName,
			Name:            hiveTable.Spec.TableName,
			Columns:         hiveTable.Spec.Columns,
			PartitionedBy:   hiveTable.Spec.PartitionedBy,
			ClusteredBy:     hiveTable.Spec.ClusteredBy,
			SortedBy:        newSortedBy,
			NumBuckets:      hiveTable.Spec.NumBuckets,
			Location:        hiveTable.Spec.Location,
			RowFormat:       hiveTable.Spec.RowFormat,
			FileFormat:      hiveTable.Spec.FileFormat,
			TableProperties: hiveTable.Spec.TableProperties,
			External:        hiveTable.Spec.External,
		}
		err := op.hiveTableManager.CreateTable(params, true)
		if err != nil {
			return fmt.Errorf("couldn't create table %s in Hive: %v", hiveTable.Spec.TableName, err)
		}

		logger.Infof("successfully created table %s in Hive", hiveTable.Spec.TableName)

		hiveTable.Status.TableName = hiveTable.Spec.TableName
		hiveTable.Status.DatabaseName = hiveTable.Spec.DatabaseName
		hiveTable.Status.Columns = hiveTable.Spec.Columns
		hiveTable.Status.PartitionedBy = hiveTable.Spec.PartitionedBy
		hiveTable.Status.ClusteredBy = hiveTable.Spec.ClusteredBy
		hiveTable.Status.SortedBy = hiveTable.Spec.SortedBy
		hiveTable.Status.NumBuckets = hiveTable.Spec.NumBuckets
		hiveTable.Status.Location = hiveTable.Spec.Location
		hiveTable.Status.RowFormat = hiveTable.Spec.RowFormat
		hiveTable.Status.FileFormat = hiveTable.Spec.FileFormat
		hiveTable.Status.TableProperties = hiveTable.Spec.TableProperties
		hiveTable.Status.External = hiveTable.Spec.External
		hiveTable.Status.Partitions = hiveTable.Spec.Partitions
		hiveTable, err = op.meteringClient.MeteringV1().HiveTables(hiveTable.Namespace).Update(hiveTable)
		if err != nil {
			return err
		}
		prestoColumns, err := reportingutil.HiveColumnsToPrestoColumns(append(hiveTable.Spec.Columns, hiveTable.Spec.PartitionedBy...))
		if err != nil {
			return fmt.Errorf("unable to create PrestoTable %s, error converting Hive columns to Presto columns: %s", hiveTable.Name, err)
		}

		ownerRef := metav1.NewControllerRef(hiveTable, metering.HiveTableGVK)
		prestoTable := &metering.PrestoTable{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PrestoTable",
				APIVersion: metering.PrestoTableGVK.GroupVersion().String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      hiveTable.Name,
				Namespace: hiveTable.Namespace,
				Labels:    hiveTable.ObjectMeta.Labels,
				OwnerReferences: []metav1.OwnerReference{
					*ownerRef,
				},
			},
			Spec: metering.PrestoTableSpec{
				// this is managed via the HiveTable, do not manage it directly
				// as a Presto table
				Unmanaged: true,
				Catalog:   "hive",
				Schema:    hiveTable.Status.DatabaseName,
				TableName: hiveTable.Status.TableName,
				Columns:   prestoColumns,
			},
		}
		prestoTable, err = op.meteringClient.MeteringV1().PrestoTables(hiveTable.Namespace).Create(prestoTable)
		if err != nil {
			if apierrors.IsAlreadyExists(err) && prestoTable.Status.TableName != "" {
				logger.Infof("PrestoTable %s already exists", prestoTable.Name)
			} ***REMOVED*** {
				return fmt.Errorf("couldn't create PrestoTable resource %s: %v", hiveTable.Name, err)
			}
		}
	}

	if hiveTable.Spec.ManagePartitions {
		currentPartitions := hiveTable.Status.Partitions
		desiredPartitions := hiveTable.Spec.Partitions
		hivePartitionColumns := hiveTable.Spec.PartitionedBy
		changes := getPartitionChanges(hivePartitionColumns, currentPartitions, desiredPartitions)

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

		// We append toUpdatePartitions to both slices because we process an
		// update by removing it ***REMOVED***rst, then adding it back with the new
		// values.
		toRemove := append(changes.toRemovePartitions, changes.toUpdatePartitions...)
		toAdd := append(changes.toAddPartitions, changes.toUpdatePartitions...)

		tableName := hiveTable.Status.TableName
		for _, p := range toRemove {
			partSpecStr := reporting.FmtPartitionSpec(hivePartitionColumns, p.PartitionSpec)
			locStr := ""
			if p.Location != "" {
				locStr = "location " + p.Location
			}
			logger.Debugf("dropping partition %s %s from Hive table %q", partSpecStr, locStr, tableName)
			err := op.hivePartitionManager.DropPartition(tableName, hivePartitionColumns, hive.TablePartition(p))
			if err != nil {
				return fmt.Errorf("failed to drop partition %s %s from Hive table %q: %s", partSpecStr, locStr, tableName, err)
			}
			logger.Debugf("partition successfully dropped partition %s %s from Hive table %q", partSpecStr, locStr, tableName)
		}

		for _, p := range toAdd {
			partSpecStr := reporting.FmtPartitionSpec(hivePartitionColumns, p.PartitionSpec)
			locStr := ""
			if p.Location != "" {
				locStr = "location " + p.Location
			}
			logger.Debugf("adding partition %s %s to Hive table %q", partSpecStr, locStr, locStr, tableName)
			err := op.hivePartitionManager.AddPartition(tableName, hivePartitionColumns, hive.TablePartition(p))
			if err != nil {
				return fmt.Errorf("failed to add partition %s %s to Hive table %q: %s", partSpecStr, locStr, tableName, err)
			}
			logger.Debugf("partition successfully added partition %s %s to Hive table %q", partSpecStr, locStr, tableName)
		}

		hiveTable.Status.Partitions = desiredPartitions
		var err error
		hiveTable, err = op.meteringClient.MeteringV1().HiveTables(hiveTable.Namespace).Update(hiveTable)
		if err != nil {
			return err
		}
		hiveTablePartitionsGauge.WithLabelValues(tableName).Set(float64(len(desiredPartitionsList)))
	}

	return nil
}

func (op *Reporting) addHiveTableFinalizer(hiveTable *metering.HiveTable) (*metering.HiveTable, error) {
	hiveTable.Finalizers = append(hiveTable.Finalizers, hiveTableFinalizer)
	newHiveTable, err := op.meteringClient.MeteringV1().HiveTables(hiveTable.Namespace).Update(hiveTable)
	logger := op.logger.WithFields(log.Fields{"hiveTable": hiveTable.Name, "namespace": hiveTable.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s ***REMOVED***nalizer to HiveTable: %s/%s", hiveTableFinalizer, hiveTable.Namespace, hiveTable.Name)
		return nil, err
	}
	logger.Infof("added %s ***REMOVED***nalizer to HiveTable: %s/%s", hiveTableFinalizer, hiveTable.Namespace, hiveTable.Name)
	return newHiveTable, nil
}

func (op *Reporting) removeHiveTableFinalizer(hiveTable *metering.HiveTable) (*metering.HiveTable, error) {
	if !slice.ContainsString(hiveTable.ObjectMeta.Finalizers, hiveTableFinalizer, nil) {
		return hiveTable, nil
	}
	hiveTable.Finalizers = slice.RemoveString(hiveTable.Finalizers, hiveTableFinalizer, nil)
	logger := op.logger.WithFields(log.Fields{"hiveTable": hiveTable.Name, "namespace": hiveTable.Namespace})
	newHiveTable, err := op.meteringClient.MeteringV1().HiveTables(hiveTable.Namespace).Update(hiveTable)
	if err != nil {
		logger.WithError(err).Errorf("error removing %s ***REMOVED***nalizer from HiveTable: %s/%s", hiveTableFinalizer, hiveTable.Namespace, hiveTable.Name)
		return nil, err
	}
	logger.Infof("removed %s ***REMOVED***nalizer from HiveTable: %s/%s", hiveTableFinalizer, hiveTable.Namespace, hiveTable.Name)
	return newHiveTable, nil
}

func hiveTableNeedsFinalizer(hiveTable *metering.HiveTable) bool {
	return hiveTable.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(hiveTable.ObjectMeta.Finalizers, hiveTableFinalizer, nil)
}

func (op *Reporting) createHiveTableCR(obj metav1.Object, gvk schema.GroupVersionKind, params hive.TableParameters, managePartitions bool, partitions []hive.TablePartition) (*metering.HiveTable, error) {
	apiVersion := gvk.GroupVersion().String()
	kind := gvk.Kind
	name := obj.GetName()
	namespace := obj.GetNamespace()
	objLabels := obj.GetLabels()
	ownerRef := metav1.NewControllerRef(obj, gvk)

	var ***REMOVED***nalizers []string
	if op.cfg.EnableFinalizers {
		***REMOVED***nalizers = []string{hiveTableFinalizer}
	}

	var newPartitions []metering.HiveTablePartition
	for _, p := range partitions {
		newPartitions = append(newPartitions, metering.HiveTablePartition(p))
	}
	var newSortedBy []metering.SortColumn
	for _, c := range params.SortedBy {
		newSortedBy = append(newSortedBy, metering.SortColumn{
			Name:      c.Name,
			Decending: c.Decending,
		})
	}

	resourceName := reportingutil.TableResourceNameFromKind(kind, namespace, name)
	newHiveTable := &metering.HiveTable{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HiveTable",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels:    objLabels,
			OwnerReferences: []metav1.OwnerReference{
				*ownerRef,
			},
			Finalizers: ***REMOVED***nalizers,
		},
		Spec: metering.HiveTableSpec{
			DatabaseName:     params.Database,
			TableName:        params.Name,
			Columns:          params.Columns,
			PartitionedBy:    params.PartitionedBy,
			ClusteredBy:      params.ClusteredBy,
			SortedBy:         newSortedBy,
			NumBuckets:       params.NumBuckets,
			Location:         params.Location,
			RowFormat:        params.RowFormat,
			TableProperties:  params.TableProperties,
			External:         params.External,
			ManagePartitions: managePartitions,
			Partitions:       newPartitions,
		},
	}
	var err error
	hiveTable, err := op.meteringClient.MeteringV1().HiveTables(namespace).Create(newHiveTable)
	switch {
	case apierrors.IsAlreadyExists(err):
		op.logger.Warnf("HiveTable %s already exists", resourceName)
		hiveTable, err = op.hiveTableLister.HiveTables(namespace).Get(resourceName)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, fmt.Errorf("couldn't create HiveTable resource for %s %s: %v", gvk, name, err)
	}
	return hiveTable, nil
}

func (op *Reporting) waitForHiveTable(namespace, name string, pollInterval, timeout time.Duration) (*metering.HiveTable, error) {
	var hiveTable *metering.HiveTable
	err := wait.Poll(pollInterval, timeout, func() (bool, error) {
		var err error
		hiveTable, err = op.meteringClient.MeteringV1().HiveTables(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if hiveTable.Status.TableName != "" {
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
	return hiveTable, nil
}

func (op *Reporting) dropHiveTable(hiveTable *metering.HiveTable) error {
	tableName := hiveTable.Status.TableName
	databaseName := hiveTable.Status.DatabaseName
	logger := op.logger.WithFields(log.Fields{"hiveTable": hiveTable.Name, "namespace": hiveTable.Namespace, "tableName": tableName})
	logger.Infof("dropping hive table %s", tableName)
	err := op.hiveTableManager.DropTable(databaseName, tableName, true)
	if err != nil {
		logger.WithError(err).Error("unable to drop hive table")
		return err
	}
	logger.Infof("successfully deleted table %s", tableName)
	return nil
}

type partitionChanges struct {
	toRemovePartitions []metering.HiveTablePartition
	toAddPartitions    []metering.HiveTablePartition
	toUpdatePartitions []metering.HiveTablePartition
}

func getPartitionChanges(partitionColumns []hive.Column, current []metering.HiveTablePartition, desired []metering.HiveTablePartition) partitionChanges {
	currSet := sets.NewString()
	desiredSet := sets.NewString()
	// lookup a map used to go back from setID to Partition
	lookup := make(map[string]metering.HiveTablePartition)

	for _, p := range current {
		var vals []string
		for _, cols := range partitionColumns {
			vals = append(vals, p.PartitionSpec[cols.Name])
		}
		setID := strings.Join(vals, "-")
		lookup[setID] = p
		currSet.Insert(setID)
	}
	for _, p := range desired {
		var vals []string
		for _, cols := range partitionColumns {
			vals = append(vals, p.PartitionSpec[cols.Name])
		}
		setID := strings.Join(vals, "-")
		lookup[setID] = p
		desiredSet.Insert(setID)
	}

	toRemove := currSet.Difference(desiredSet)
	toAdd := desiredSet.Difference(currSet)
	toUpdate := currSet.Intersection(desiredSet)

	var changes partitionChanges
	for _, setID := range toRemove.UnsortedList() {
		p := lookup[setID]
		changes.toRemovePartitions = append(changes.toRemovePartitions, p)
	}
	for _, setID := range toAdd.UnsortedList() {
		p := lookup[setID]
		changes.toAddPartitions = append(changes.toAddPartitions, p)
	}
	for _, setID := range toUpdate.UnsortedList() {
		p := lookup[setID]
		changes.toUpdatePartitions = append(changes.toUpdatePartitions, p)
	}

	return changes
}

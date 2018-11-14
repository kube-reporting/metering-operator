package operator

import (
	"fmt"
	"net/url"
	"path"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

func (op *Reporting) createTableForStorage(logger log.FieldLogger, obj metav1.Object, gvk schema.GroupVersionKind, storage *cbTypes.StorageLocationRef, tableName string, columns, partitions []hive.Column) error {
	tableProperties, err := op.getHiveTableProperties(logger, storage, gvk.Kind)
	if err != nil {
		return fmt.Errorf("storage incorrectly con***REMOVED***gured for %s %s, err: %v", gvk, obj.GetName(), err)
	}
	tableParams := hive.TableParameters{
		Name:         tableName,
		Columns:      columns,
		Partitions:   partitions,
		IgnoreExists: true,
	}
	return op.createTableWith(logger, obj, gvk, tableParams, *tableProperties)
}

func (op *Reporting) createTableForStorageNoCR(logger log.FieldLogger, storage *cbTypes.StorageLocationRef, tableName string, columns []hive.Column) error {
	tableProperties, err := op.getHiveTableProperties(logger, storage, tableName)
	if err != nil {
		return fmt.Errorf("storage incorrectly con***REMOVED***gured for %s, err: %v", tableName, err)
	}
	tableParams := hive.TableParameters{
		Name:         tableName,
		Columns:      columns,
		IgnoreExists: true,
	}
	newTableProperties, err := addTableNameToLocation(*tableProperties, tableName)
	if err != nil {
		return err
	}
	return op.createTable(logger, tableParams, newTableProperties)
}

func (op *Reporting) createTableWith(logger log.FieldLogger, obj metav1.Object, gvk schema.GroupVersionKind, params hive.TableParameters, properties hive.TableProperties) error {
	newTableProperties, err := addTableNameToLocation(properties, params.Name)
	if err != nil {
		return err
	}
	return op.createTableAndCR(logger, obj, gvk, params, newTableProperties)
}

func (op *Reporting) createTableAndCR(logger log.FieldLogger, obj metav1.Object, gvk schema.GroupVersionKind, params hive.TableParameters, properties hive.TableProperties) error {
	err := op.createTable(logger, params, properties)
	if err != nil {
		return err
	}
	err = op.createPrestoTableCR(obj, gvk, params, properties, nil)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Infof("presto table resource already exists")
		} ***REMOVED*** {
			return fmt.Errorf("couldn't create PrestoTable resource for %s %s: %v", gvk, obj.GetName(), err)
		}
	}
	return nil
}

func (op *Reporting) createTable(logger log.FieldLogger, params hive.TableParameters, properties hive.TableProperties) error {
	logger.Debugf("Creating table %s with Hive Storage %#v", params.Name, properties)
	err := op.tableManager.CreateTable(params, properties)
	if err != nil {
		return fmt.Errorf("couldn't create table: %v", err)
	}
	logger.Debugf("successfully created table %s", params.Name)
	return nil
}

func addTableNameToLocation(tableProperties hive.TableProperties, tableName string) (hive.TableProperties, error) {
	// Validate the URL
	u, err := url.Parse(tableProperties.Location)
	if err != nil {
		return tableProperties, err
	}
	// Append the tableName to the location; as tables shouldn't have
	// overlapping locations.
	u.Path = path.Join(u.Path, tableName)
	tableProperties.Location = u.String()
	return tableProperties, nil
}

package chargeback

import (
	"fmt"
	"net/url"
	"path"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

func (c *Chargeback) createTableForStorage(logger log.FieldLogger, obj runtime.Object, kind, name string, storage *cbTypes.StorageLocationRef, tableName string, columns []hive.Column) error {
	tableProperties, err := c.getHiveTableProperties(logger, storage, kind)
	if err != nil {
		return fmt.Errorf("storage incorrectly configured for %s: %s", kind, name)
	}
	tableParams := hive.TableParameters{
		Name:         tableName,
		Columns:      columns,
		IgnoreExists: true,
	}
	return c.createTableWith(logger, obj, kind, name, tableParams, *tableProperties)
}

func (c *Chargeback) createTableForStorageNoCR(logger log.FieldLogger, storage *cbTypes.StorageLocationRef, tableName string, columns []hive.Column) error {
	tableProperties, err := c.getHiveTableProperties(logger, storage, tableName)
	if err != nil {
		return fmt.Errorf("storage incorrectly configured for %s", tableName)
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
	return c.createTable(logger, tableParams, newTableProperties)
}

func (c *Chargeback) createTableWith(logger log.FieldLogger, obj runtime.Object, kind, name string, params hive.TableParameters, properties hive.TableProperties) error {
	newTableProperties, err := addTableNameToLocation(properties, params.Name)
	if err != nil {
		return err
	}
	return c.createTableAndCR(logger, obj, kind, name, params, newTableProperties)
}

func (c *Chargeback) createTableAndCR(logger log.FieldLogger, obj runtime.Object, kind, name string, params hive.TableParameters, properties hive.TableProperties) error {
	err := c.createTable(logger, params, properties)
	if err != nil {
		return err
	}
	err = c.createPrestoTableCR(obj, cbTypes.GroupName, kind, params, properties, nil)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Infof("presto table resource already exists")
		} else {
			return fmt.Errorf("couldn't create PrestoTable resource for %s: %v", kind, err)
		}
	}
	return nil
}

func (c *Chargeback) createTable(logger log.FieldLogger, params hive.TableParameters, properties hive.TableProperties) error {
	logger.Debugf("Creating table %s with Hive Storage %#v", params.Name, properties)
	err := hive.ExecuteCreateTable(c.hiveQueryer, params, properties)
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

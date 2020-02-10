package reporting

import (
	"fmt"
	"strings"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

type HiveTableManager interface {
	CreateTable(params hive.TableParameters, ignoreExists bool) error
	DropTable(dbName, tableName string, ignoreNotExists bool) error
}

type HiveDatabaseManager interface {
	CreateDatabase(params hive.DatabaseParameters) error
	DropDatabase(dbName string, ignoreNotExists, cascade bool) error
}

type HivePartitionManager interface {
	AddPartition(dbName, tableName string, partitionColumns []hive.Column, partition hive.TablePartition) error
	DropPartition(dbName, tableName string, partitionColumns []hive.Column, partition hive.TablePartition) error
}

type HiveManager struct {
	execer db.Execer
}

func NewHiveManager(execer db.Execer) *HiveManager {
	return &HiveManager{execer: execer}
}

func (m *HiveManager) CreateTable(params hive.TableParameters, ignoreExists bool) error {
	return hive.ExecuteCreateTable(m.execer, params, ignoreExists)
}

func (m *HiveManager) DropTable(dbName, tableName string, ignoreNotExists bool) error {
	return hive.ExecuteDropTable(m.execer, dbName, tableName, ignoreNotExists)
}

func (m *HiveManager) CreateDatabase(params hive.DatabaseParameters) error {
	return hive.ExecuteCreateDatabase(m.execer, params)
}

func (m *HiveManager) DropDatabase(dbName string, ignoreNotExists, cascade bool) error {
	return hive.ExecuteDropDatabase(m.execer, dbName, ignoreNotExists, cascade)
}

func (m *HiveManager) AddPartition(dbName, tableName string, partitionColumns []hive.Column, partition hive.TablePartition) error {
	partitionSpecStr := FmtPartitionSpec(partitionColumns, partition.PartitionSpec)
	locationStr := ""
	if partition.Location != "" {
		locationStr = fmt.Sprintf("LOCATION '%s'", partition.Location)
	}
	_, err := m.execer.Exec(fmt.Sprintf("ALTER TABLE %s.%s ADD IF NOT EXISTS PARTITION (%s) %s", dbName, tableName, partitionSpecStr, locationStr))
	return err
}

func (m *HiveManager) DropPartition(dbName, tableName string, partitionColumns []hive.Column, partition hive.TablePartition) error {
	partitionSpecStr := FmtPartitionSpec(partitionColumns, partition.PartitionSpec)
	_, err := m.execer.Exec(fmt.Sprintf("ALTER TABLE %s.%s DROP IF EXISTS PARTITION (%s)", dbName, tableName, partitionSpecStr))
	return err
}

func FmtPartitionSpec(partitionColumns []hive.Column, partSpec hive.PartitionSpec) string {
	var partitionVals []string
	for _, col := range partitionColumns {
		val := partSpec[col.Name]
		// Quote strings
		if strings.ToLower(col.Type) == "string" {
			val = fmt.Sprintf("'%s'", val)
		}
		partitionVals = append(partitionVals, fmt.Sprintf("`%s`=%s", col.Name, val))
	}
	return strings.Join(partitionVals, ", ")
}

type PrestoTableManager interface {
	CreateTable(catalog, schema, tableName string, columns []presto.Column, comment string, properties map[string]string, ignoreExists bool) error
	CreateTableAs(catalog, schema, tableName string, columns []presto.Column, comment string, properties map[string]string, ignoreExists bool, query string) error
	DropTable(catalog, schema, tableName string, ignoreNotExists bool) error
	QueryMetadata(catalog, schema, tableName string) ([]presto.Column, error)

	CreateView(catalog, schema, viewName, query string) error
	DropView(catalog, schema, viewName string, ignoreNotExists bool) error
}

type PrestoTableManagerImpl struct {
	queryer db.Queryer
}

func NewPrestoTableManager(queryer db.Queryer) *PrestoTableManagerImpl {
	return &PrestoTableManagerImpl{queryer: queryer}
}

func (c *PrestoTableManagerImpl) CreateTable(catalog, schema, tableName string, columns []presto.Column, comment string, properties map[string]string, ignoreExists bool) error {
	return presto.CreateTable(c.queryer, catalog, schema, tableName, columns, comment, properties, ignoreExists)
}

func (c *PrestoTableManagerImpl) CreateTableAs(catalog, schema, tableName string, columns []presto.Column, comment string, properties map[string]string, ignoreExists bool, query string) error {
	return presto.CreateTableAs(c.queryer, catalog, schema, tableName, columns, comment, properties, ignoreExists, query)
}

func (c *PrestoTableManagerImpl) DropTable(catalog, schema, tableName string, ignoreNotExists bool) error {
	return presto.DropTable(c.queryer, catalog, schema, tableName, ignoreNotExists)
}

func (c *PrestoTableManagerImpl) CreateView(catalog, schema, viewName, query string) error {
	return presto.CreateView(c.queryer, catalog, schema, viewName, query, true)
}

func (c *PrestoTableManagerImpl) DropView(catalog, schema, viewName string, ignoreNotExists bool) error {
	return presto.DropView(c.queryer, catalog, schema, viewName, ignoreNotExists)
}

func (c *PrestoTableManagerImpl) QueryMetadata(catalog, schema, tableName string) ([]presto.Column, error) {
	return presto.QueryMetadata(c.queryer, catalog, schema, tableName)
}

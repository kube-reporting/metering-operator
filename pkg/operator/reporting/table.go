package reporting

import (
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
)

type TableManager interface {
	CreateTable(params hive.TableParameters, properties hive.TableProperties) error
	DropTable(tableName string, ignoreNotExists bool) error
}

type AWSTablePartitionManager interface {
	AddPartition(tableName, start, end, location string) error
	DropPartition(tableName, start, end string) error
}

type HiveTableManager struct {
	queryer db.Queryer
}

func NewHiveTableManager(queryer db.Queryer) *HiveTableManager {
	return &HiveTableManager{queryer: queryer}
}

func (m *HiveTableManager) CreateTable(params hive.TableParameters, properties hive.TableProperties) error {
	return hive.ExecuteCreateTable(m.queryer, params, properties)
}

func (m *HiveTableManager) DropTable(tableName string, ignoreNotExists bool) error {
	return hive.ExecuteDropTable(m.queryer, tableName, ignoreNotExists)
}

func (m *HiveTableManager) AddPartition(tableName, start, end, location string) error {
	return reportingutil.AddAWSHivePartition(m.queryer, tableName, start, end, location)
}

func (m *HiveTableManager) DropPartition(tableName, start, end string) error {
	return reportingutil.DropAWSHivePartition(m.queryer, tableName, start, end)
}

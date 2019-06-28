package framework

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
)

func (f *Framework) GetPrestoTable(name string) (*metering.PrestoTable, error) {
	return f.MeteringClient.PrestoTables(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForPrestoTable(t *testing.T, name string, pollInterval, timeout time.Duration, tableFunc func(table *metering.PrestoTable) (bool, error)) (*metering.PrestoTable, error) {
	t.Helper()
	var table *metering.PrestoTable
	return table, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		var err error
		table, err = f.GetPrestoTable(name)
		if err != nil {
			if errors.IsNotFound(err) {
				t.Logf("PrestoTable %s does not exist yet", name)
				return false, nil
			}
			return false, err
		}
		return tableFunc(table)
	})
}

func (f *Framework) PrestoTableExists(t *testing.T, name string) (bool, error) {
	prestoTable, err := f.MeteringClient.PrestoTables(f.Namespace).Get(name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			t.Logf("PrestoTable %s resource does not exist yet", name)
			return false, nil
		}
		return false, err
	}

	if prestoTable.Status.TableName == "" {
		t.Logf("PrestoTable %s status.tableName not set yet", prestoTable.Name)
		return false, nil
	}

	t.Logf("PrestoTable %s has a table: %s", prestoTable.Name, prestoTable.Status.TableName)
	return true, nil
}

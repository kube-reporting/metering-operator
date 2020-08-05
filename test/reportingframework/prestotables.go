package reportingframework

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
)

// GetPrestoTable is a reportingframework method that makes the
// client-go API request for the @name PrestoTable.
func (rf *ReportingFramework) GetPrestoTable(name string) (*metering.PrestoTable, error) {
	return rf.MeteringClient.PrestoTables(rf.Namespace).Get(context.Background(), name, metav1.GetOptions{})
}

// WaitForPrestoTable is a reportingframework method responsbile
// for ensuring that the @name PrestoTable custom resource in the
// rf.Namespace namespace has is reporting a ready status. We define
// "ready" here based on the @tableFunc anonymous function parameter.
func (rf *ReportingFramework) WaitForPrestoTable(t *testing.T, name string, pollInterval, timeout time.Duration, tableFunc func(table *metering.PrestoTable) (bool, error)) (*metering.PrestoTable, error) {
	t.Helper()

	var table *metering.PrestoTable
	return table, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		var err error
		table, err = rf.GetPrestoTable(name)
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

// PrestoTableExists is a reportingframework method that determines
// whether or not the @name PrestoTable custom resource in the
// rf.Namespace namespace has a populated database table in Presto.
func (rf *ReportingFramework) PrestoTableExists(t *testing.T, name string) (bool, error) {
	prestoTable, err := rf.MeteringClient.PrestoTables(rf.Namespace).Get(context.Background(), name, metav1.GetOptions{})
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

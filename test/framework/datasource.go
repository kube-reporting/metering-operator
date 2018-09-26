package framework

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (f *Framework) GetMeteringReportDataSource(name string) (*meteringv1alpha1.ReportDataSource, error) {
	return f.MeteringClient.ReportDataSources(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportDataSourceTable(t *testing.T, name string, pollInterval, timeout time.Duration) (*meteringv1alpha1.ReportDataSource, error) {
	var ds *meteringv1alpha1.ReportDataSource
	return ds, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		var err error
		ds, err = f.GetMeteringReportDataSource(name)
		if err != nil {
			if errors.IsNotFound(err) {
				t.Logf("ReportDataSource %s does not exist yet", name)
				return false, nil
			}
			return false, err
		}
		if ds.Status.TableName == "" {
			t.Logf("ReportDataSource %s table is not created yet", name)
			return false, nil
		}
		return true, nil
	})
}

func (f *Framework) WaitForAllMeteringReportDataSourceTables(t *testing.T, pollInterval, timeout time.Duration) ([]*meteringv1alpha1.ReportDataSource, error) {
	var reportDataSources []*meteringv1alpha1.ReportDataSource
	return reportDataSources, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		reportDataSourcesList, err := f.MeteringClient.ReportDataSources(f.Namespace).List(meta.ListOptions{})
		require.NoError(t, err, "should not have errors querying API for list of ReportDataSources")
		reportDataSources = reportDataSourcesList.Items

		for _, ds := range reportDataSources {
			if ds.Status.TableName == "" {
				t.Logf("ReportDataSource %s table is not created yet", ds.Name)
				return false, nil
			}
		}
		return true, nil
	})
}

package framework

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (f *Framework) GetMeteringReportDataSource(name string) (*metering.ReportDataSource, error) {
	return f.MeteringClient.ReportDataSources(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportDataSourceTable(t *testing.T, name string, pollInterval, timeout time.Duration) (*metering.ReportDataSource, error) {
	t.Helper()
	ds, err := f.WaitForMeteringReportDataSource(t, name, pollInterval, timeout, func(ds *metering.ReportDataSource) (bool, error) {
		if ds.Status.TableRef.Name == "" {
			t.Logf("ReportDataSource %s table is not created yet", name)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return nil, fmt.Errorf("timed out waiting for ReportDataSource %s table", name)
		}
		return nil, err
	}
	return ds, nil
}

func (f *Framework) WaitForAllMeteringReportDataSourceTables(t *testing.T, pollInterval, timeout time.Duration) ([]*metering.ReportDataSource, error) {
	t.Helper()
	var reportDataSources []*metering.ReportDataSource
	return reportDataSources, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		reportDataSourcesList, err := f.MeteringClient.ReportDataSources(f.Namespace).List(meta.ListOptions{})
		require.NoError(t, err, "should not have errors querying API for list of ReportDataSources")
		reportDataSources = reportDataSourcesList.Items

		for _, ds := range reportDataSources {
			if ds.Status.TableRef.Name == "" {
				t.Logf("ReportDataSource %s table is not created yet", ds.Name)
				return false, nil
			}
		}
		return true, nil
	})
}

func (f *Framework) WaitForMeteringReportDataSource(t *testing.T, name string, pollInterval, timeout time.Duration, dsFunc func(ds *metering.ReportDataSource) (bool, error)) (*metering.ReportDataSource, error) {
	t.Helper()
	var ds *metering.ReportDataSource
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
		return dsFunc(ds)
	})
}

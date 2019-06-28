package framework

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (f *Framework) GetMeteringReportQuery(name string) (*metering.ReportQuery, error) {
	return f.MeteringClient.ReportQueries(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportQuery(t *testing.T, name string, pollInterval, timeout time.Duration) (*metering.ReportQuery, error) {
	t.Helper()
	var reportQuery *metering.ReportQuery
	return reportQuery, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		var err error
		reportQuery, err = f.GetMeteringReportQuery(name)
		if err != nil {
			if errors.IsNotFound(err) {
				t.Logf("ReportQuery %s does not exist yet", name)
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
}

func (f *Framework) RequireReportQueriesReady(t *testing.T, queries []string, pollInterval, timeout time.Duration) {
	t.Helper()
	readyReportDataSources := make(map[string]struct{})
	readyReportGenQueries := make(map[string]struct{})

	reportGetter := reporting.NewReportClientGetter(f.MeteringClient)
	queryGetter := reporting.NewReportQueryClientGetter(f.MeteringClient)
	dataSourceGetter := reporting.NewReportDataSourceClientGetter(f.MeteringClient)

	for _, queryName := range queries {
		if _, exists := readyReportGenQueries[queryName]; exists {
			continue
		}

		t.Logf("waiting for ReportQuery %s to exist", queryName)
		reportQuery, err := f.WaitForMeteringReportQuery(t, queryName, pollInterval, timeout)
		require.NoError(t, err, "ReportQuery should exist before creating report using it")

		depHandler := &reporting.UninitialiedDependendenciesHandler{
			HandleUninitializedReportDataSource: func(ds *metering.ReportDataSource) {
				if _, exists := readyReportDataSources[ds.Name]; exists {
					return
				}
				t.Logf("%s dependencies: waiting for ReportDataSource %s to exist", queryName, ds.Name)
				_, err := f.WaitForMeteringReportDataSourceTable(t, ds.Name, pollInterval, timeout)
				require.NoError(t, err, "ReportDataSource %s table for ReportQuery %s should exist before running reports against it", ds.Name, queryName)
				readyReportDataSources[ds.Name] = struct{}{}
			},
		}

		t.Logf("waiting for ReportQuery %s dependencies to become initialized", queryName)
		// explicitly ignoring results, since we'll get errors above if any of
		// the uninitialized dependencies don't become ready in the handler
		_, _ = reporting.GetAndValidateQueryDependencies(queryGetter, dataSourceGetter, reportGetter, reportQuery, nil, depHandler)
		readyReportGenQueries[queryName] = struct{}{}
	}
}

func (f *Framework) RequireReportDataSourcesForQueryHaveData(t *testing.T, queries []string, collectResp operator.CollectPrometheusMetricsDataResponse) {
	t.Helper()
	reportGetter := reporting.NewReportClientGetter(f.MeteringClient)
	queryGetter := reporting.NewReportQueryClientGetter(f.MeteringClient)
	dataSourceGetter := reporting.NewReportDataSourceClientGetter(f.MeteringClient)

	metricsImportedForDS := make(map[string]int)
	for _, res := range collectResp.Results {
		metricsImportedForDS[res.ReportDataSource] = res.MetricsImportedCount
	}

	for _, queryName := range queries {
		query, err := f.GetMeteringReportQuery(queryName)
		require.NoError(t, err, "ReportQuery should exist")
		deps, err := reporting.GetQueryDependencies(queryGetter, dataSourceGetter, reportGetter, query, nil)
		require.NoError(t, err, "Getting ReportQuery dependencies should succeed")

		for _, dataSource := range deps.ReportDataSources {
			metricsImported := metricsImportedForDS[dataSource.Name]
			require.NotZerof(t, metricsImported, "expected metric import count for ReportDataSource %s to not be zero", dataSource.Name)
		}
	}
}

package framework

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (f *Framework) GetMeteringReportGenerationQuery(name string) (*metering.ReportGenerationQuery, error) {
	return f.MeteringClient.ReportGenerationQueries(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportGenerationQuery(t *testing.T, name string, pollInterval, timeout time.Duration) (*metering.ReportGenerationQuery, error) {
	var reportQuery *metering.ReportGenerationQuery
	return reportQuery, wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		var err error
		reportQuery, err = f.GetMeteringReportGenerationQuery(name)
		if err != nil {
			if errors.IsNotFound(err) {
				t.Logf("ReportGenerationQuery %s does not exist yet", name)
				return false, nil
			}
			return false, err
		}
		if !reportQuery.Spec.View.Disabled && reportQuery.Status.ViewName == "" {
			t.Logf("ReportGenerationQuery %s view is not created yet", name)
			return false, nil
		}
		return true, nil
	})
}

func (f *Framework) RequireReportGenerationQueriesReady(t *testing.T, queries []string, pollInterval, timeout time.Duration) {
	readyReportDataSources := make(map[string]struct{})
	readyReportGenQueries := make(map[string]struct{})

	reportGetter := reporting.NewReportClientGetter(f.MeteringClient)
	scheduledReportGetter := reporting.NewScheduledReportClientGetter(f.MeteringClient)
	queryGetter := reporting.NewReportGenerationQueryClientGetter(f.MeteringClient)
	dataSourceGetter := reporting.NewReportDataSourceClientGetter(f.MeteringClient)

	for _, queryName := range queries {
		if _, exists := readyReportGenQueries[queryName]; exists {
			continue
		}

		t.Logf("waiting for ReportGenerationQuery %s to exist", queryName)
		reportGenQuery, err := f.WaitForMeteringReportGenerationQuery(t, queryName, pollInterval, timeout)
		require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")

		depHandler := &reporting.UninitialiedDependendenciesHandler{
			HandleUninitializedReportGenerationQuery: func(query *metering.ReportGenerationQuery) {
				if _, exists := readyReportGenQueries[query.Name]; exists {
					return
				}
				t.Logf("%s dependencies: waiting for ReportGenerationQuery %s to exist", queryName, query.Name)
				_, err := f.WaitForMeteringReportGenerationQuery(t, query.Name, pollInterval, timeout)
				require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")
				readyReportGenQueries[query.Name] = struct{}{}
			},
			HandleUninitializedReportDataSource: func(ds *metering.ReportDataSource) {
				if _, exists := readyReportDataSources[ds.Name]; exists {
					return
				}
				t.Logf("%s dependencies: waiting for ReportDataSource %s to exist", queryName, ds.Name)
				_, err := f.WaitForMeteringReportDataSourceTable(t, ds.Name, pollInterval, timeout)
				require.NoError(t, err, "ReportDataSource %s table for ReportGenerationQuery %s should exist before running reports against it", ds.Name, queryName)
				readyReportDataSources[ds.Name] = struct{}{}
			},
		}

		t.Logf("waiting for ReportGenerationQuery %s dependencies to become initialized", queryName)
		// explicitly ignoring results, since we'll get errors above if any of
		// the uninitialized dependencies don't become ready in the handler
		_, _ = reporting.GetAndValidateGenerationQueryDependencies(queryGetter, dataSourceGetter, reportGetter, scheduledReportGetter, reportGenQuery, depHandler)
		readyReportGenQueries[queryName] = struct{}{}
	}
}

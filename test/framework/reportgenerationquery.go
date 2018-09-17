package framework

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/stretchr/testify/require"
)

func (f *Framework) GetMeteringReportGenerationQuery(name string) (*meteringv1alpha1.ReportGenerationQuery, error) {
	return f.MeteringClient.ReportGenerationQueries(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportGenerationQuery(t *testing.T, name string, pollInterval, timeout time.Duration) (*meteringv1alpha1.ReportGenerationQuery, error) {
	var reportQuery *meteringv1alpha1.ReportGenerationQuery
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
		return true, nil
	})
}

func (f *Framework) RequireReportGenerationQueriesReady(t *testing.T, queries []string, pollInterval, timeout time.Duration) {
	readyReportDataSources := make(map[string]struct{})
	readyReportGenQueries := make(map[string]struct{})

	queryGetter := reporting.NewReportGenerationQueryClientGetter(f.MeteringClient)
	dataSourceGetter := reporting.NewReportDataSourceClientGetter(f.MeteringClient)

	for _, queryName := range queries {
		if _, exists := readyReportGenQueries[queryName]; exists {
			continue
		}

		t.Logf("waiting for ReportGenerationQuery %s to exist", queryName)
		reportGenQuery, err := f.WaitForMeteringReportGenerationQuery(t, queryName, pollInterval, timeout)
		require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")

		depStatus, err := reporting.GetGenerationQueryDependenciesStatus(queryGetter, dataSourceGetter, reportGenQuery)
		require.NoError(t, err, "should not have errors getting dependent ReportGenerationQueries")

		var uninitializedReportGenerationQueries, uninitializedReportDataSources []string

		for _, q := range depStatus.UninitializedReportGenerationQueries {
			uninitializedReportGenerationQueries = append(uninitializedReportGenerationQueries, q.Name)
		}

		for _, ds := range depStatus.UninitializedReportDataSources {
			uninitializedReportDataSources = append(uninitializedReportDataSources, ds.Name)
		}

		t.Logf("waiting for ReportGenerationQuery %s UninitializedReportGenerationQueries: %v", queryName, uninitializedReportGenerationQueries)
		t.Logf("waiting for ReportGenerationQuery %s UninitializedReportDataSources: %v", queryName, uninitializedReportDataSources)

		for _, q := range depStatus.UninitializedReportGenerationQueries {
			if _, exists := readyReportGenQueries[q.Name]; exists {
				continue
			}

			t.Logf("waiting for ReportGenerationQuery %s to exist", q.Name)
			_, err := f.WaitForMeteringReportGenerationQuery(t, q.Name, pollInterval, timeout)
			require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")

			readyReportGenQueries[queryName] = struct{}{}
		}

		for _, ds := range depStatus.UninitializedReportDataSources {
			if _, exists := readyReportDataSources[ds.Name]; exists {
				continue
			}
			t.Logf("waiting for ReportDataSource %s to exist", ds.Name)
			_, err := f.WaitForMeteringReportDataSourceTable(t, ds.Name, pollInterval, timeout)
			require.NoError(t, err, "ReportDataSource %s table for ReportGenerationQuery %s should exist before running reports against it", ds.Name, queryName)
			readyReportDataSources[ds.Name] = struct{}{}
		}

		readyReportGenQueries[queryName] = struct{}{}
	}
}

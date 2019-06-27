package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/test/framework"
)

type reportProducesDataTestCase struct {
	name          string
	queryName     string
	schedule      *metering.ReportSchedule
	newReportFunc func(name, queryName string, schedule *metering.ReportSchedule, start, end *time.Time) *metering.Report
	skip          bool
	parallel      bool
}

func testReportsProduceData(t *testing.T, testFramework *framework.Framework, testCases []reportProducesDataTestCase) {
	for _, test := range testCases {
		name := test.name
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip("test configured to be skipped")
				return
			}
			if test.parallel {
				t.Parallel()
			}

			query, err := testFramework.GetMeteringReportQuery(test.queryName)
			require.NoError(t, err, "report query for report should exist")

			dsGetter := reporting.NewReportDataSourceClientGetter(testFramework.MeteringClient)
			queryGetter := reporting.NewReportQueryClientGetter(testFramework.MeteringClient)
			reportGetter := reporting.NewReportClientGetter(testFramework.MeteringClient)

			// get all the datasources for the query used in our report
			dependencies, err := reporting.GetQueryDependencies(queryGetter, dsGetter, reportGetter, query, nil)
			require.NoError(t, err, "datasources for query should exist")

			require.NotEqual(t, 0, len(dependencies.ReportDataSources), "Report should have at least 1 datasource dependency")

			var reportStart time.Time

			// for each datasource, wait until it's EarliestImportedMetricTime is set
			for _, ds := range dependencies.ReportDataSources {
				_, err := testFramework.WaitForMeteringReportDataSource(t, ds.Name, 5*time.Second, 5*time.Minute, func(dataSource *metering.ReportDataSource) (bool, error) {
					if dataSource.Spec.PrometheusMetricsImporter == nil {
						return true, nil
					}
					if dataSource.Status.PrometheusMetricsImportStatus != nil && dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime != nil {
						// keep the EarliestImportedMetricTime that is the
						// least far back, so that we ensure the reportStart is
						// a time that all datasources have metrics for.
						if reportStart.IsZero() || dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime.Time.After(reportStart) {
							reportStart = dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime.Time
						}
						return true, nil
					}
					return false, nil
				})
				require.NoError(t, err, "expected ReportDataSource %s to have an earliestImportedMetricTime", ds.Name)
			}

			if reportStart.IsZero() {
				t.Errorf("reportStart is zero")
			}

			reportStart = reportStart.UTC()
			// The report spans 5 minutes
			reportEnd := reportStart.Add(5 * time.Minute).UTC()

			t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

			report := test.newReportFunc(test.name, test.queryName, test.schedule, &reportStart, &reportEnd)
			reportRunTimeout := 10 * time.Minute
			t.Logf("creating report %s and waiting %s to finish", report.Name, reportRunTimeout)
			testFramework.RequireReportSuccessfullyRuns(t, report, reportRunTimeout)

			resultTimeout := time.Minute
			t.Logf("waiting %s for report %s results", resultTimeout, report.Name)
			reportResults := testFramework.GetReportResults(t, report, resultTimeout)
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")
		})
	}
}

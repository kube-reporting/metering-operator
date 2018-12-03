package integration

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/test/framework"
)

var testFramework *framework.Framework

func TestMain(m *testing.M) {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	ns := flag.String("namespace", "metering-ci", "test namespace")
	httpsAPI := flag.Bool("https-api", false, "If true, use https to talk to Metering API")
	flag.Parse()

	var err error
	if testFramework, err = framework.New(*ns, *kubeconfig, *httpsAPI); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}

func TestReportingIntegration(t *testing.T) {
	t.Run("TestReportsProduceCorrectData", func(t *testing.T) {
		var queries []string
		waitTimeout := time.Minute

		_, err := testFramework.WaitForAllMeteringReportDataSourceTables(t, time.Second*5, waitTimeout)
		require.NoError(t, err, "should not error when waiting for all ReportDataSource tables to be created")

		for _, test := range testReportsProduceCorrectDataForInputTestCases {
			queries = append(queries, test.queryName)
		}

		// validate all ReportGenerationQueries and ReportDataSources that are
		// used by the test cases are initialized
		testFramework.RequireReportGenerationQueriesReady(t, queries, time.Second*5, waitTimeout)

		var reportStart, reportEnd time.Time
		dataSourcesSubmitted := make(map[string]struct{})

		for _, test := range testReportsProduceCorrectDataForInputTestCases {
			for _, dataSource := range test.dataSources {
				if _, alreadySubmitted := dataSourcesSubmitted[dataSource.DatasourceName]; !alreadySubmitted {
					// wait for the datasource table to exist
					_, err := testFramework.WaitForMeteringReportDataSourceTable(t, dataSource.DatasourceName, time.Second*5, test.timeout)
					require.NoError(t, err, "ReportDataSource table should exist before storing data into it")

					metricsFile, err := os.Open(dataSource.FileName)
					require.NoError(t, err)

					decoder := json.NewDecoder(metricsFile)

					_, err = decoder.Token()
					require.NoError(t, err)

					var metrics []*prestostore.PrometheusMetric
					for decoder.More() {
						var metric prestostore.PrometheusMetric
						err = decoder.Decode(&metric)
						require.NoError(t, err)
						if reportStart.IsZero() || metric.Timestamp.Before(reportStart) {
							reportStart = metric.Timestamp
						}
						if metric.Timestamp.After(reportEnd) {
							reportEnd = metric.Timestamp
						}
						metrics = append(metrics, &metric)
						// batch store metrics in amounts of 100
						if len(metrics) >= 100 {
							err := testFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
							require.NoError(t, err)
							metrics = nil
						}
					}
					// flush any metrics left over
					if len(metrics) != 0 {
						err = testFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
						require.NoError(t, err)
					}

					dataSourcesSubmitted[dataSource.DatasourceName] = struct{}{}
				}
			}
		}

		t.Run("TestReportsProduceCorrectDataForInput", func(t *testing.T) {
			testReportsProduceCorrectDataForInput(t, reportStart, reportEnd, testReportsProduceCorrectDataForInputTestCases)
		})
	})
}

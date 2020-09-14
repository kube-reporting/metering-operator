package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/cache"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	listers "github.com/kube-reporting/metering-operator/pkg/generated/listers/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator/prestostore"
	"github.com/kube-reporting/metering-operator/pkg/presto"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

var (
	testRandSeed               = rand.NewSource(0)
	testRand                   = rand.New(testRandSeed)
	noopPrometheusImporterFunc = func(ctx context.Context, namespace, dsName string, start, end time.Time) ([]*prometheusImportResults, error) {
		return nil, nil
	}
	testLogger = logrus.New()
)

//for v2 endpoints full
func apiReportV2URLFull(namespace, reportName string) string {
	return path.Join(APIV2ReportEndpointPrefix, namespace, reportName, "full")
}

//for v2 endpoints TableHidden
func apiReportV2URLTable(namespace, reportName string) string {
	return path.Join(APIV2ReportEndpointPrefix, namespace, reportName, "table")
}

type fakePrometheusMetricsRepo struct {
	metrics map[string][]*prestostore.PrometheusMetric
	err     error
}

func (f *fakePrometheusMetricsRepo) StorePrometheusMetrics(ctx context.Context, tableName string, metrics []*prestostore.PrometheusMetric) error {
	if f.err != nil {
		return f.err
	}
	f.metrics[tableName] = append(f.metrics[tableName], metrics...)

	// sort metrics we store by timestamp
	sort.Slice(f.metrics[tableName], func(i, j int) bool {
		return f.metrics[tableName][i].Timestamp.Before(f.metrics[tableName][j].Timestamp)
	})
	return nil
}

func (f *fakePrometheusMetricsRepo) GetPrometheusMetrics(tableName string, start, end time.Time) ([]*prestostore.PrometheusMetric, error) {
	if f.err != nil {
		return nil, f.err
	}
	if metrics, ok := f.metrics[tableName]; ok {
		return metrics, nil
	}
	return nil, fmt.Errorf("table %s not found", tableName)
}

func (f *fakePrometheusMetricsRepo) GetLastTimestampForTable(tableName string) (*time.Time, error) {
	if metrics, ok := f.metrics[tableName]; ok {
		return &metrics[len(metrics)-1].Timestamp, nil
	}
	return nil, fmt.Errorf("table %s not found", tableName)
}

type fakeReportResultsGetter struct {
	results []presto.Row
	err     error
}

func (f *fakeReportResultsGetter) GetReportResults(tableName string, columns []presto.Column) ([]presto.Row, error) {
	return f.results, f.err
}

func TestAPIV1ReportsGet(t *testing.T) {
	const (
		namespace       = "default"
		testReportName  = "test-report"
		testQueryName   = "test-query"
		testCatalogName = "hive"
		testSchemaName  = "default"
	)
	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp

	tests := map[string]struct {
		reportName string

		report      *metering.Report
		query       *metering.ReportQuery
		prestoTable *metering.PrestoTable

		expectedStatusCode int
		expectedAPIError   string
		expectedResults    []presto.Row

		prometheusMetricsRepo prestostore.PrometheusMetricsRepo
		reportResultsGetter   prestostore.ReportResultsGetter
	}{
		"report-finished-no-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			expectedResults:       nil,
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusOK,
		},
		"report-finished-with-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			},
			),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			expectedResults: []presto.Row{
				{
					"timestamp": time.Time{},
					"foo":       1.5,
				},
			},
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusOK,
		},
		"report-finished-db-errored": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				results: nil,
				err:     errors.New("mock database had an error"),
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "failed to perform presto query",
		},
		"non-existent-report": {
			reportName:            "doesnt-exist",
			expectedStatusCode:    http.StatusNotFound,
			expectedAPIError:      "not found",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-name-not-specified": {
			reportName:            "",
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "the following fields are missing or empty: name",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"mismatched-results-schema-to-table-schema": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			expectedResults: []presto.Row{
				{
					"timestamp": time.Time{},
					"foo":       1.5,
					"this_column_doesnt_exist_in_presto_table": "fail",
				},
			},
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
						"this_column_doesnt_exist_in_presto_table": "fail",
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "results schema doesn't match expected schema",
		},
	}

	for testName, tt := range tests {
		tt := tt
		testName := testName
		t.Run(testName, func(t *testing.T) {
			// manually create indexers which implement the lister interfaces
			// needed for meteringListers. Since a cache.Indexer is a
			// cache.Store, it's basically just a key-value store that we can
			// use to mock the lister returns.
			reportIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportQueryIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportDataSourceIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			prestoTableIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

			reportLister := listers.NewReportLister(reportIndexer)
			reportQueryLister := listers.NewReportQueryLister(reportQueryIndexer)
			reportDataSourceLister := listers.NewReportDataSourceLister(reportDataSourceIndexer)
			prestoTableLister := listers.NewPrestoTableLister(prestoTableIndexer)

			// add our test report if one is specified
			if tt.report != nil {
				reportIndexer.Add(tt.report)
			}
			// add our test query for the report
			if tt.query != nil {
				reportQueryIndexer.Add(tt.query)
			}
			if tt.prestoTable != nil {
				prestoTableIndexer.Add(tt.prestoTable)
			}

			// setup a test server suitable for making API calls against
			router := newRouter(testLogger, testRand, tt.prometheusMetricsRepo, tt.reportResultsGetter, nil, noopPrometheusImporterFunc,
				reportLister, reportDataSourceLister, reportQueryLister, prestoTableLister,
			)
			server := httptest.NewServer(router)
			defer server.Close()

			// set up the query parameters for our API call.
			// we hardcode format to JSON because validating CSV output is a
			// bit trickier than JSON
			params := url.Values{
				"format":    []string{"json"},
				"name":      []string{tt.reportName},
				"namespace": []string{namespace},
			}

			endpoint := server.URL + APIV1ReportGetEndpoint

			// construct the url object
			endpointURL, err := url.Parse(endpoint)
			require.NoError(t, err)
			endpointURL.RawQuery = params.Encode()

			// final string URL
			finalURL := endpointURL.String()

			// query the API
			resp, err := server.Client().Get(finalURL)
			require.NoError(t, err, "expected making http request to not return error")

			// read the body from the response. it shouldn't be empty, even if
			// the response had a non-2xx code
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err, "expected read all of resp.Body to succeed")

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected http status code to match")
			t.Logf("response body: %s", string(body))

			// if tt.expectedAPIError is non-empty, then this test is testing the
			// error response from the API, otherwise it's testing the results
			// come back with the correct number of items
			if tt.expectedAPIError != "" {
				var errResp errorResponse
				err = json.Unmarshal(body, &errResp)
				assert.NoError(t, err, "expected unmarshal to not error")
				assert.Contains(t, errResp.Error, tt.expectedAPIError, "expected error response to contain expected api error")
			} else {
				var results []presto.Row
				err = json.Unmarshal(body, &results)
				assert.NoError(t, err, "expected unmarshal to not error")
				// TODO(chance): check more than the results length matching
				assert.Len(t, results, len(tt.expectedResults), "expected API results length to match expected results length")
			}
		})
	}
}

func TestAPIV2ReportsFull(t *testing.T) {
	const (
		namespace       = "default"
		testReportName  = "test-report"
		testQueryName   = "test-query"
		testFormat      = "?format=json"
		testCatalogName = "hive"
		testSchemaName  = "default"
	)
	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp

	tests := map[string]struct {
		reportStatus metering.ReportStatus
		reportName   string
		reportFormat string
		apiPath      string

		report      *metering.Report
		query       *metering.ReportQuery
		prestoTable *metering.PrestoTable

		expectedStatusCode int
		expectedAPIError   string
		expectedResults    *GetReportResults

		prometheusMetricsRepo prestostore.PrometheusMetricsRepo
		reportResultsGetter   prestostore.ReportResultsGetter
	}{
		"report-finished-with-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLFull(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1,
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusOK,
			expectedResults: &GetReportResults{
				Results: []ReportResultEntry{
					{
						Values: []ReportResultValues{
							{
								Name:        "foo",
								Value:       1,
								TableHidden: false,
							},
						},
					},
				},
			},
		},
		"report-finished-no-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLFull(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedResults:       &GetReportResults{},
			expectedStatusCode:    http.StatusOK,
		},
		"report-finished-db-errored": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLFull(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				err: errors.New("mock database had an error"),
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "failed to perform presto query",
		},
		"non-existent-report": {
			reportName:            "doesnt-exist",
			apiPath:               apiReportV2URLFull(namespace, "doesnt-exist") + testFormat,
			expectedStatusCode:    http.StatusNotFound,
			expectedAPIError:      "not found",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-name-not-specified": {
			reportName:            "",
			apiPath:               APIV2ReportEndpointPrefix + "/ " + namespace + "//full" + testFormat,
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "the following fields are missing or empty: name",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-format-not-specified": {
			reportName:            testReportName,
			report:                testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:               apiReportV2URLFull(namespace, testReportName),
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "the following fields are missing or empty: format",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-format-non-existent": {
			reportName:            testReportName,
			report:                testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:               apiReportV2URLFull(namespace, testReportName) + "?format=doesntexist",
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "format must be one of: csv, json or tabular",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"mismatched-results-schema-to-table-schema": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLFull(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
						"this_column_doesnt_exist_in_presto_table": "fail",
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "results schema doesn't match expected schema",
		},
	}

	for testName, tt := range tests {
		tt := tt
		testName := testName
		t.Run(testName, func(t *testing.T) {
			// manually create indexers which implement the lister interfaces
			// needed for meteringListers. Since a cache.Indexer is a
			// cache.Store, it's basically just a key-value store that we can
			// use to mock the lister returns.
			reportIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportQueryIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportDataSourceIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			prestoTableIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

			reportLister := listers.NewReportLister(reportIndexer)
			reportQueryLister := listers.NewReportQueryLister(reportQueryIndexer)
			reportDataSourceLister := listers.NewReportDataSourceLister(reportDataSourceIndexer)
			prestoTableLister := listers.NewPrestoTableLister(prestoTableIndexer)

			// add our test report if one is specified
			if tt.report != nil {
				reportIndexer.Add(tt.report)
			}
			// add our test query for the report
			if tt.query != nil {
				reportQueryIndexer.Add(tt.query)
			}
			if tt.prestoTable != nil {
				prestoTableIndexer.Add(tt.prestoTable)
			}

			// setup a test server suitable for making API calls against
			router := newRouter(testLogger, testRand, tt.prometheusMetricsRepo, tt.reportResultsGetter, nil, noopPrometheusImporterFunc,
				reportLister, reportDataSourceLister, reportQueryLister, prestoTableLister,
			)
			server := httptest.NewServer(router)
			defer server.Close()

			// final string URL
			finalURL := server.URL + tt.apiPath
			resp, err := server.Client().Get(finalURL)
			require.NoError(t, err, "expected making http request to not return error")

			// read the body from the response. it shouldn't be empty, even if
			// the response had a non-2xx code
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err, "expected read all of resp.Body to succeed")

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected http status code to match")
			t.Logf("response body: %s", string(body))

			// if tt.expectedAPIError is non-empty, then this test is testing the
			// error response from the API, otherwise it's testing the results
			// come back with the correct number of items
			if tt.expectedAPIError != "" {
				var errResp errorResponse
				err = json.Unmarshal(body, &errResp)
				assert.NoError(t, err, "expected unmarshal to not error")
				assert.Contains(t, errResp.Error, tt.expectedAPIError, "expected error response to contain expected api error")
			} else {
				var results GetReportResults
				err = json.Unmarshal(body, &results)
				assert.NoError(t, err, "expected unmarshal to not error")
				// TODO(chance): check more than the results length matching
				assert.Len(t, results.Results, len(tt.expectedResults.Results), "expected API results length to match expected results length")
			}
		})
	}
}

func TestAPIV2ReportsTable(t *testing.T) {
	const (
		namespace       = "default"
		testReportName  = "test-report"
		testQueryName   = "test-query"
		testFormat      = "?format=json"
		testCatalogName = "hive"
		testSchemaName  = "default"
	)
	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp

	tests := map[string]struct {
		reportStatus metering.ReportStatus
		reportName   string
		reportFormat string
		apiPath      string

		report      *metering.Report
		query       *metering.ReportQuery
		prestoTable *metering.PrestoTable

		expectedStatusCode int
		expectedAPIError   string
		expectedResults    *GetReportResults

		prometheusMetricsRepo prestostore.PrometheusMetricsRepo
		reportResultsGetter   prestostore.ReportResultsGetter
	}{
		"report-finished-with-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLTable(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			},
			),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			},
			),
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1,
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusOK,
			expectedResults: &GetReportResults{
				Results: []ReportResultEntry{
					{
						Values: []ReportResultValues{
							{
								Name:        "foo",
								Value:       1,
								TableHidden: false,
							},
						},
					},
				},
			},
		},
		"report-finished-no-results": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLTable(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedResults:       &GetReportResults{},
			expectedStatusCode:    http.StatusOK,
		},
		"report-finished-db-errored": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLTable(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				err: errors.New("mock database had an error"),
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "failed to perform presto query",
		},
		"non-existent-report": {
			reportName:            "doesnt-exist",
			apiPath:               apiReportV2URLTable(namespace, "doesnt-exist") + testFormat,
			expectedStatusCode:    http.StatusNotFound,
			expectedAPIError:      "not found",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-name-not-specified": {
			reportName:            "",
			apiPath:               APIV2ReportEndpointPrefix + "/ " + namespace + "//table" + testFormat,
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "the following fields are missing or empty: name",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-format-not-specified": {
			reportName:            testReportName,
			report:                testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:               apiReportV2URLTable(namespace, testReportName),
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "the following fields are missing or empty: format",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"report-format-non-existent": {
			reportName:            testReportName,
			report:                testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:               apiReportV2URLTable(namespace, testReportName) + "?format=doesntexist",
			expectedStatusCode:    http.StatusBadRequest,
			expectedAPIError:      "format must be one of: csv, json or tabular",
			reportResultsGetter:   &fakeReportResultsGetter{},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
		},
		"mismatched-results-schema-to-table-schema": {
			reportName: testReportName,
			report:     testhelpers.NewReport(testReportName, namespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			apiPath:    apiReportV2URLTable(namespace, testReportName) + testFormat,
			query: testhelpers.NewReportQuery(testQueryName, namespace, []metering.ReportQueryColumn{
				{
					Name:        "timestamp",
					Type:        "timestamp",
					TableHidden: true,
				},
				{
					Name:        "foo",
					Type:        "double",
					TableHidden: false,
				},
			}),
			prestoTable: testhelpers.NewPrestoTable(testReportName, namespace, testCatalogName, testSchemaName, []presto.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			reportResultsGetter: &fakeReportResultsGetter{
				results: []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
						"this_column_doesnt_exist_in_presto_table": "fail",
					},
				},
			},
			prometheusMetricsRepo: &fakePrometheusMetricsRepo{},
			expectedStatusCode:    http.StatusInternalServerError,
			expectedAPIError:      "results schema doesn't match expected schema",
		},
	}

	for testName, tt := range tests {
		tt := tt
		testName := testName
		t.Run(testName, func(t *testing.T) {
			// manually create indexers which implement the lister interfaces
			// needed for meteringListers. Since a cache.Indexer is a
			// cache.Store, it's basically just a key-value store that we can
			// use to mock the lister returns.
			reportIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportQueryIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			reportDataSourceIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			prestoTableIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

			reportLister := listers.NewReportLister(reportIndexer)
			reportQueryLister := listers.NewReportQueryLister(reportQueryIndexer)
			reportDataSourceLister := listers.NewReportDataSourceLister(reportDataSourceIndexer)
			prestoTableLister := listers.NewPrestoTableLister(prestoTableIndexer)

			// add our test report if one is specified
			if tt.report != nil {
				reportIndexer.Add(tt.report)
			}
			// add our test query for the report
			if tt.query != nil {
				reportQueryIndexer.Add(tt.query)
			}
			if tt.prestoTable != nil {
				prestoTableIndexer.Add(tt.prestoTable)
			}

			// setup a test server suitable for making API calls against
			router := newRouter(testLogger, testRand, tt.prometheusMetricsRepo, tt.reportResultsGetter, nil, noopPrometheusImporterFunc,
				reportLister, reportDataSourceLister, reportQueryLister, prestoTableLister,
			)
			server := httptest.NewServer(router)
			defer server.Close()

			// final string URL
			finalURL := server.URL + tt.apiPath
			resp, err := server.Client().Get(finalURL)
			require.NoError(t, err, "expected making http request to not return error")

			// read the body from the response. it shouldn't be empty, even if
			// the response had a non-2xx code
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err, "expected read all of resp.Body to succeed")

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected http status code to match")
			t.Logf("response body: %s", string(body))

			// if tt.expectedAPIError is non-empty, then this test is testing the
			// error response from the API, otherwise it's testing the results
			// come back with the correct number of items
			if tt.expectedAPIError != "" {
				var errResp errorResponse
				err = json.Unmarshal(body, &errResp)
				assert.NoError(t, err, "expected unmarshal to not error")
				assert.Contains(t, errResp.Error, tt.expectedAPIError, "expected error response to contain expected api error")
			} else {
				var results GetReportResults
				err = json.Unmarshal(body, &results)
				assert.NoError(t, err, "expected unmarshal to not error")
				// TODO(chance): check more than the results length matching
				assert.Len(t, results.Results, len(tt.expectedResults.Results), "expected API results length to match expected results length")
			}
		})
	}
}

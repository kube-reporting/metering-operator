package chargeback

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	listers "github.com/operator-framework/operator-metering/pkg/generated/listers/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/presto/mock"
)

var (
	testRandSeed               = rand.NewSource(0)
	testRand                   = rand.New(testRandSeed)
	noopPrometheusImporterFunc = func(ctx context.Context, start, end time.Time) error {
		return nil
	}
	testLogger = logrus.New()
)

func newTestReport(name, namespace, testQueryName string, reportStart, reportEnd time.Time, reportStatus v1alpha1.ReportStatus) *v1alpha1.Report {
	return &v1alpha1.Report{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReportSpec{
			GenerationQueryName: testQueryName,
			ReportingStart:      meta.Time{reportStart},
			ReportingEnd:        meta.Time{reportEnd},
			RunImmediately:      true,
		},
		Status: reportStatus,
	}
}

func newTestReportGenQuery(name, namespace string, columns []v1alpha1.ReportGenerationQueryColumn) *v1alpha1.ReportGenerationQuery {
	return &v1alpha1.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReportGenerationQuerySpec{
			Columns: columns,
		},
	}
}

func newTestPrestoTable(name, namespace string, columns []hive.Column) *v1alpha1.PrestoTable {
	return &v1alpha1.PrestoTable{
		ObjectMeta: meta.ObjectMeta{
			Name:      prestoTableResourceNameFromKind("report", name),
			Namespace: namespace,
		},
		State: v1alpha1.PrestoTableState{
			Parameters: v1alpha1.TableParameters{
				Columns: columns,
			},
		},
	}
}

func TestAPIV1ReportsGet(t *testing.T) {
	const namespace = "default"
	const testReportName = "test-report"
	const testQueryName = "test-query"
	reportStart := time.Time{}
	reportEnd := reportStart.AddDate(0, 1, 0)

	tests := map[string]struct {
		reportName string

		report      *v1alpha1.Report
		query       *v1alpha1.ReportGenerationQuery
		prestoTable *v1alpha1.PrestoTable

		expectedStatusCode int
		expectedAPIError   string

		queryerPrepareFunc func(*mockpresto.MockExecQueryer) []presto.Row
	}{
		"report-***REMOVED***nished-no-results": {
			reportName: testReportName,
			report:     newTestReport(testReportName, namespace, testQueryName, reportStart, reportEnd, v1alpha1.ReportStatus{Phase: v1alpha1.ReportPhaseFinished}),
			query: newTestReportGenQuery(testQueryName, namespace, []v1alpha1.ReportGenerationQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: newTestPrestoTable(testReportName, namespace, []hive.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			queryerPrepareFunc: func(mock *mockpresto.MockExecQueryer) []presto.Row {
				mock.EXPECT().Query(gomock.Any()).Return(nil, nil)
				return nil
			},
			expectedStatusCode: http.StatusOK,
		},
		"report-***REMOVED***nished-with-results": {
			reportName: testReportName,
			report:     newTestReport(testReportName, namespace, testQueryName, reportStart, reportEnd, v1alpha1.ReportStatus{Phase: v1alpha1.ReportPhaseFinished}),
			query: newTestReportGenQuery(testQueryName, namespace, []v1alpha1.ReportGenerationQueryColumn{
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
			prestoTable: newTestPrestoTable(testReportName, namespace, []hive.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			queryerPrepareFunc: func(mock *mockpresto.MockExecQueryer) []presto.Row {
				result := []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
					},
				}
				mock.EXPECT().Query(gomock.Any()).Return(result, nil)
				return result
			},
			expectedStatusCode: http.StatusOK,
		},
		"report-***REMOVED***nished-db-errored": {
			reportName: testReportName,
			report:     newTestReport(testReportName, namespace, testQueryName, reportStart, reportEnd, v1alpha1.ReportStatus{Phase: v1alpha1.ReportPhaseFinished}),
			query: newTestReportGenQuery(testQueryName, namespace, []v1alpha1.ReportGenerationQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: newTestPrestoTable(testReportName, namespace, []hive.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			queryerPrepareFunc: func(mock *mockpresto.MockExecQueryer) []presto.Row {
				dbErr := errors.New("mock database had an error")
				mock.EXPECT().Query(gomock.Any()).Return(nil, dbErr)
				return nil
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedAPIError:   "failed to perform presto query",
		},
		"non-existent-report": {
			reportName:         "doesnt-exist",
			expectedStatusCode: http.StatusNotFound,
			expectedAPIError:   "not found",
		},
		"mismatched-results-schema-to-table-schema": {
			reportName: testReportName,
			report:     newTestReport(testReportName, namespace, testQueryName, reportStart, reportEnd, v1alpha1.ReportStatus{Phase: v1alpha1.ReportPhaseFinished}),
			query: newTestReportGenQuery(testQueryName, namespace, []v1alpha1.ReportGenerationQueryColumn{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			prestoTable: newTestPrestoTable(testReportName, namespace, []hive.Column{
				{
					Name: "timestamp",
					Type: "timestamp",
				},
				{
					Name: "foo",
					Type: "double",
				},
			}),
			queryerPrepareFunc: func(mock *mockpresto.MockExecQueryer) []presto.Row {
				result := []presto.Row{
					{
						"timestamp": time.Time{},
						"foo":       1.5,
						"this_column_doesnt_exist_in_presto_table": "fail",
					},
				}
				mock.EXPECT().Query(gomock.Any()).Return(result, nil)
				return result
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedAPIError:   "results schema doesn't match expected schema",
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
			reportGenerationQueryIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})
			prestoTableIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

			listers := meteringListers{
				reports:                 listers.NewReportLister(reportIndexer).Reports(namespace),
				reportGenerationQueries: listers.NewReportGenerationQueryLister(reportGenerationQueryIndexer).ReportGenerationQueries(namespace),
				prestoTables:            listers.NewPrestoTableLister(prestoTableIndexer).PrestoTables(namespace),
			}

			// add our test report if one is speci***REMOVED***ed
			if tt.report != nil {
				reportIndexer.Add(tt.report)
			}
			// add our test query for the report
			if tt.query != nil {
				reportGenerationQueryIndexer.Add(tt.query)
			}
			if tt.prestoTable != nil {
				prestoTableIndexer.Add(tt.prestoTable)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// setup our mock queryer so we can test without a real database
			queryer := mockpresto.NewMockExecQueryer(ctrl)

			// expectedResults is what our mock queryer will return. since the
			// v1 results endpoint just serializes the slice of rows returned
			// from the DB, this is also what we expect the HTTP api to
			// return when there are no errors
			var expectedResults []presto.Row
			if tt.queryerPrepareFunc != nil {
				expectedResults = tt.queryerPrepareFunc(queryer)
			}

			// setup a test server suitable for making API calls against
			router := newRouter(testLogger, queryer, testRand, noopPrometheusImporterFunc, listers)
			server := httptest.NewServer(router)
			defer server.Close()

			// set up the query parameters for our API call.
			// we hardcode format to JSON because validating CSV output is a
			// bit trickier than JSON
			params := url.Values{
				"format": []string{"json"},
				"name":   []string{tt.reportName},
			}
			endpoint := server.URL + APIV1ReportsGetEndpoint

			// construct the url object
			endpointURL, err := url.Parse(endpoint)
			require.NoError(t, err)
			endpointURL.RawQuery = params.Encode()

			// ***REMOVED***nal string URL
			***REMOVED***nalURL := endpointURL.String()

			// query the API
			resp, err := server.Client().Get(***REMOVED***nalURL)
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
			} ***REMOVED*** {
				var results []presto.Row
				err = json.Unmarshal(body, &results)
				assert.NoError(t, err, "expected unmarshal to not error")
				// TODO(chance): check more than the results length matching
				assert.Len(t, results, len(expectedResults), "expected API results length to match expected results length")
			}
		})
	}
}

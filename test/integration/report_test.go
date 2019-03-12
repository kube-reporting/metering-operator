package integration

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/test/testhelpers"
	"github.com/stretchr/testify/require"
)

type reportsProduceCorrectDataForInputTestCase struct {
	name                         string
	queryName                    string
	dataSources                  []testDatasource
	expectedReportOutputFileName string
	comparisonColumnNames        []string
	timeout                      time.Duration
	parallel                     bool
}

func testReportsProduceCorrectDataForInput(t *testing.T, reportStart, reportEnd time.Time, testCases []reportsProduceCorrectDataForInputTestCase) {
	require.NotZero(t, reportStart, "reportStart should not be zero")
	require.NotZero(t, reportEnd, "reportEnd should not be zero")
	t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)
	for _, test := range testCases {
		// Fix closure captures
		name := test.name
		test := test
		t.Run(name, func(t *testing.T) {
			if test.parallel {
				t.Parallel()
			}

			report := testFramework.NewSimpleReport(test.name+"-runonce", test.queryName, nil, &reportStart, &reportEnd)

			reportRunTimeout := 10 * time.Minute
			t.Logf("creating report %s and waiting %s to ***REMOVED***nish", report.Name, reportRunTimeout)
			testFramework.RequireReportSuccessfullyRuns(t, report, reportRunTimeout)

			resultTimeout := time.Minute
			t.Logf("waiting %s for report %s results", resultTimeout, report.Name)
			actualResults := testFramework.GetReportResults(t, report, resultTimeout)

			// read expected results from a ***REMOVED***le
			expectedReportData, err := ioutil.ReadFile(test.expectedReportOutputFileName)
			require.NoError(t, err)
			// turn the expected results into a list of maps
			var expectedResults []map[string]interface{}
			err = json.Unmarshal(expectedReportData, &expectedResults)
			require.NoError(t, err)

			testhelpers.AssertReportResultsEqual(t, expectedResults, actualResults, test.comparisonColumnNames)
		})
	}
}

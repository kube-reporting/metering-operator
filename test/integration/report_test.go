package integration

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testReportsProduceCorrectDataForInput(t *testing.T, reportStart, reportEnd time.Time, testCases []reportsProduceCorrectDataForInputTestCase) {
	require.NotZero(t, reportStart, "reportStart should not be zero")
	require.NotZero(t, reportEnd, "reportEnd should not be zero")

	for _, test := range testCases {
		// Fix closure captures
		name := test.name
		test := test
		t.Run(name, func(t *testing.T) {

			t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

			report := testFramework.NewSimpleReport(test.name, test.queryName, nil, &reportStart, &reportEnd)

			defer func() {
				t.Logf("deleting scheduled report %s", report.Name)
				err := testFramework.MeteringClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete scheduled report to succeed")
			}()

			testFramework.RequireReportSuccessfullyRuns(t, report, time.Minute)

			// read expected results from a ***REMOVED***le
			expectedReportData, err := ioutil.ReadFile(test.expectedReportOutputFileName)
			require.NoError(t, err)
			// turn the expected results into a list of maps
			var expectedResults []map[string]interface{}
			err = json.Unmarshal(expectedReportData, &expectedResults)
			require.NoError(t, err)

			actualResults := testFramework.GetReportResults(t, report, time.Minute)

			testhelpers.AssertReportResultsEqual(t, expectedResults, actualResults, test.comparisonColumnNames)
		})
	}
}

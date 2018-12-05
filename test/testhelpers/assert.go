package testhelpers

import (
	"testing"

	"github.com/operator-framework/operator-metering/pkg/util/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reportComparisionEpsilon = 0.0001

func AssertReportResultsEqual(t *testing.T, expected, actual []map[string]interface{}, comparisonColumnNames []string) {
	// turn the list of expected results maps into a list of ordered maps
	expectedResults := make([]*orderedmap.OrderedMap, len(expected))
	for i, item := range expected {
		var err error
		expectedResults[i], err = orderedmap.NewFromMap(item)
		require.NoError(t, err)
	}

	// turn the list of actual results maps into a list of ordered maps
	actualResults := make([]*orderedmap.OrderedMap, len(actual))
	for i, item := range actual {
		var err error
		actualResults[i], err = orderedmap.NewFromMap(item)
		require.NoError(t, err)
	}

	require.Len(t, expectedResults, len(actualResults), "new should have same number of rows as existing report")

	// now that we have a slice of ordered maps, we should be able to
	// iterate over each row, and for each row, iterate over all
	// columns/keys in the row ensuring they match.
	// if the column is the comparison  column, then we allow a small
	// error, due to floating point precision
	// in summary, this does an deep equal comparison with a few tweaks
	// to allow for small error in the calculations.
	for i, actualRow := range actualResults {
		expectedRow := expectedResults[i]

		actualColumns := actualRow.Keys()
		expectedColumns := expectedRow.Keys()

		assert.Equal(t, actualColumns, expectedColumns, "expecting key iteration between actual and expected to be the same")
		for _, column := range actualColumns {

			actualValue, actualExists := actualRow.Get(column)
			if !actualExists {
				t.Errorf("missing column %s value from actual row", column)
			}
			expectedValue, expectedExists := expectedRow.Get(column)
			if !expectedExists {
				t.Errorf("missing column %s value from expected row", column)
			}
			isCompareColumn := false
			for _, comparisionColumn := range comparisonColumnNames {
				if comparisionColumn == column {
					isCompareColumn = true
					break
				}
			}
			if isCompareColumn {
				assert.InEpsilonf(t, actualValue, expectedValue, reportComparisionEpsilon, "expected column %q value to be within delta of expected row", column)
			} ***REMOVED*** {
				assert.Equal(t, actualValue, expectedValue, "expected column values between actual and expected rows to be the same")
			}
		}
	}
}

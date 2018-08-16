package operator

import (
	"fmt"
	"testing"

	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/stretchr/testify/assert"
)

func TestHiveColumnToPrestoColumn(t *testing.T) {
	tests := map[string]struct {
		hiveColumn           hive.Column
		expectedPrestoColumn presto.Column
		expectedErr          error
		compareAssertion     assert.ComparisonAssertionFunc
		errAssertion         assert.ErrorAssertionFunc
	}{
		"TIMESTAMP to TIMESTAMP": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "TIMESTAMP",
			},
			expectedPrestoColumn: presto.Column{
				Name: "foo",
				Type: "TIMESTAMP",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.NoError,
		},
		"STRING to VARCHAR": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "STRING",
			},
			expectedPrestoColumn: presto.Column{
				Name: "foo",
				Type: "VARCHAR",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.NoError,
		},
		"MAP<STRING,STRING> to map(VARCHAR,VARCHAR)": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<STRING,STRING>",
			},
			expectedPrestoColumn: presto.Column{
				Name: "foo",
				Type: "map(VARCHAR,VARCHAR)",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.NoError,
		},
		// taken from node-memory-allocatable, the space needs to be handled
		"map<string, string> to map(VARCHAR,VARCHAR)": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "map<string, string>",
			},
			expectedPrestoColumn: presto.Column{
				Name: "foo",
				Type: "map(VARCHAR,VARCHAR)",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.NoError,
		},
		"broken map MAP<> error": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<>",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.Error,
			expectedErr:      fmt.Errorf(`invalid map de***REMOVED***nition in column type, column "foo", type: "MAP<>"`),
		},
		"broken map MAP< error": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.Error,
			expectedErr:      fmt.Errorf(`unable to ***REMOVED***nd matching <, > pair for column "foo", type: "MAP<"`),
		},
		"broken map MAP<STRING> error": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<STRING>",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.Error,
			expectedErr:      fmt.Errorf(`invalid map de***REMOVED***nition in column type, column "foo", type: "MAP<STRING>"`),
		},
		"broken map MAP<STRING,> error": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<STRING,>",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.Error,
			expectedErr:      fmt.Errorf(`invalid presto map value type: ""`),
		},
		"broken map MAP<,STRING> error": {
			hiveColumn: hive.Column{
				Name: "foo",
				Type: "MAP<,STRING>",
			},
			compareAssertion: assert.Equal,
			errAssertion:     assert.Error,
			expectedErr:      fmt.Errorf(`invalid presto map key type: ""`),
		},
	}

	for testName, tt := range tests {
		testName := testName
		tt := tt
		t.Run(testName, func(t *testing.T) {
			prestoColumn, err := hiveColumnToPrestoColumn(tt.hiveColumn)
			tt.errAssertion(t, err)
			if tt.expectedErr != nil {
				tt.compareAssertion(t, tt.expectedErr, err)
			} ***REMOVED*** {
				tt.compareAssertion(t, tt.expectedPrestoColumn, prestoColumn)
			}
		})
	}
}

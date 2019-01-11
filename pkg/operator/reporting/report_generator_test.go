package reporting

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	mockprestostore "github.com/operator-framework/operator-metering/pkg/operator/prestostore/mock"
)

func TestGenerateReport(t *testing.T) {
	testSQL := "SELECT 1"
	testQuery := metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-query-1",
			Namespace: "default",
		},
		Spec: metering.ReportGenerationQuerySpec{
			Query: testSQL,
		},
	}
	tableName := "test-table"

	testQueryEmptyQueryField := testQuery
	testQueryEmptyQueryField.Spec.Query = ""

	// missing a bracket
	testQueryInvalidQuery := testQuery
	testQueryInvalidQuery.Spec.Query = "SELECT foo FROM {|"

	tests := map[string]struct {
		tableName                      string
		reportStart                    *time.Time
		reportEnd                      *time.Time
		reportGenerationQuery          *metering.ReportGenerationQuery
		dynamicReportGenerationQueries []*metering.ReportGenerationQuery
		inputs                         []metering.ReportGenerationQueryInputValue
		deleteExistingData             bool

		expectedErr string
	}{
		"a table name and a ReportGenerationQuery with a query ***REMOVED***eld set will succeed": {
			tableName:             tableName,
			reportGenerationQuery: &testQuery,
		},
		"an empty table name will error": {
			tableName:             "",
			reportGenerationQuery: &testQuery,
			expectedErr:           errInvalidTableName.Error(),
		},
		"an empty ReportGenerationQuery spec.query ***REMOVED***eld will error": {
			tableName:             tableName,
			reportGenerationQuery: &testQueryEmptyQueryField,
			expectedErr:           errEmptyQueryField.Error(),
		},

		"an ReportGenerationQuery spec.query with invalid template expressions will error": {
			tableName:             tableName,
			reportGenerationQuery: &testQueryInvalidQuery,
			expectedErr:           "error parsing query: template: report-generation-query:1: unexpected unclosed action in command",
		},

		"a table name and a ReportGenerationQuery with a query ***REMOVED***eld and deleteExistingData=true will succeed": {
			tableName:             tableName,
			reportGenerationQuery: &testQuery,
			deleteExistingData:    true,
		},
	}

	for testName, tt := range tests {
		testName := testName
		tt := tt
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			reportResultsRepo := mockprestostore.NewMockReportResultsRepo(ctrl)
			if tt.deleteExistingData {
				reportResultsRepo.EXPECT().DeleteReportResults(tt.tableName).Return(nil)
			}
			if tt.expectedErr == "" {
				reportResultsRepo.EXPECT().StoreReportResults(tt.tableName, tt.reportGenerationQuery.Spec.Query).Return(nil)
			}

			reportGenerator := NewReportGenerator(logger, reportResultsRepo)
			err := reportGenerator.GenerateReport(tt.tableName, "test-ns", tt.reportStart, tt.reportEnd, tt.reportGenerationQuery, tt.dynamicReportGenerationQueries, tt.inputs, tt.deleteExistingData)
			if tt.expectedErr == "" {
				assert.NoError(t, err, "expected GenerateReport to not error")
			} ***REMOVED*** {
				assert.EqualError(t, err, tt.expectedErr, "expected GenerateReport to error")
			}
		})
	}
}

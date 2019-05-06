package reporting

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	mockprestostore "github.com/operator-framework/operator-metering/pkg/operator/prestostore/mock"
)

func TestGenerateReport(t *testing.T) {
	testSQL := "SELECT 1"
	tableName := "test-table"

	tests := map[string]struct {
		tableName          string
		query              string
		deleteExistingData bool
		expectedErr        string
	}{
		"a table name and a ReportGenerationQuery with a query": {
			tableName: tableName,
			query:     testSQL,
		},
		"an empty table name will error": {
			tableName:   "",
			query:       testSQL,
			expectedErr: errInvalidTableName.Error(),
		},
		"a table name and a ReportGenerationQuery with a query field and deleteExistingData=true will succeed": {
			tableName:          tableName,
			query:              testSQL,
			deleteExistingData: true,
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
				reportResultsRepo.EXPECT().StoreReportResults(tt.tableName, tt.query).Return(nil)
			}

			reportGenerator := NewReportGenerator(logger, reportResultsRepo)
			err := reportGenerator.GenerateReport(tt.tableName, tt.query, tt.deleteExistingData)
			if tt.expectedErr == "" {
				assert.NoError(t, err, "expected GenerateReport to not error")
			} else {
				assert.EqualError(t, err, tt.expectedErr, "expected GenerateReport to error")
			}
		})
	}
}

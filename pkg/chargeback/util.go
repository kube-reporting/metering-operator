package chargeback

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/sirupsen/logrus"
)

const logIdentifierLength = 10

var resourceNameReplacer = strings.NewReplacer("-", "_", ".", "_")

func dataSourceTableName(dataSourceName string) string {
	return fmt.Sprintf("datasource_%s", resourceNameReplacer.Replace(dataSourceName))
}

func reportTableName(reportName string) string {
	return fmt.Sprintf("report_%s", resourceNameReplacer.Replace(reportName))
}

func scheduledReportTableName(reportName string) string {
	return fmt.Sprintf("scheduled_report_%s", resourceNameReplacer.Replace(reportName))
}

func generationQueryViewName(queryName string) string {
	return fmt.Sprintf("view_%s", resourceNameReplacer.Replace(queryName))
}

func prestoTableResourceNameFromKind(kind, name string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", kind, name))
}

func billingPeriodTimestamp(date time.Time) string {
	return date.Format(awsUsagePartitionDateStringLayout)
}

func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func generateHiveColumns(genQuery *cbTypes.ReportGenerationQuery) []hive.Column {
	var columns []hive.Column
	for _, c := range genQuery.Spec.Columns {
		columns = append(columns, hive.Column{Name: c.Name, Type: c.Type})
	}
	return columns
}

func generatePrestoColumns(genQuery *cbTypes.ReportGenerationQuery) []presto.Column {
	var columns []presto.Column
	for _, c := range genQuery.Spec.Columns {
		columns = append(columns, presto.Column{Name: c.Name, Type: c.Type})
	}
	return columns
}

func randomString(rand *rand.Rand, size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func newLogIdentifier(r *rand.Rand) logrus.Fields {
	return logrus.Fields{
		"logID": randomString(r, logIdentifierLength),
	}
}

func newRequestLogger(logger logrus.FieldLogger, r *http.Request, rand *rand.Rand) logrus.FieldLogger {
	return logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).WithFields(newLogIdentifier(rand))
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeErrorResponse(logger logrus.FieldLogger, w http.ResponseWriter, r *http.Request, status int, message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	writeResponseWithBody(logger, w, status, errorResponse{Error: msg})
}

// writeResponseWithBody attempts to marshal an arbitrary thing to JSON then write
// it to the http.ResponseWriter
func writeResponseWithBody(logger logrus.FieldLogger, w http.ResponseWriter, code int, resp interface{}) {
	enc, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Error("failed JSON-encoding HTTP response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(enc); err != nil {
		logger.WithError(err).Error("failed writing HTTP response")
	}
}

func validateHour(hour int64) error {
	if hour >= 0 && hour <= 23 {
		return nil
	}
	return fmt.Errorf("invalid hour: %d, must be between 0 and 23", hour)
}

func validateMinute(minute int64) error {
	if minute >= 0 && minute <= 59 {
		return nil
	}
	return fmt.Errorf("invalid minute: %d, must be between 0 and 59", minute)
}

func validateSecond(second int64) error {
	if second >= 0 && second <= 59 {
		return nil
	}
	return fmt.Errorf("invalid second: %d, must be between 0 and 59", second)
}

func validateDayOfMonth(dom int64) error {
	if dom >= 1 && dom <= 31 {
		return nil
	}
	return fmt.Errorf("invalid day of month: %d, must be between 1 and 31", dom)
}

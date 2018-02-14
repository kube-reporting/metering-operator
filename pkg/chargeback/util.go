package chargeback

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/sirupsen/logrus"
)

const logIdenti***REMOVED***erLength = 10

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

func billingPeriodFormat(date time.Time) string {
	return date.Format(hive.HiveDateStringLayout)
}

func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func generateHiveColumns(genQuery *cbTypes.ReportGenerationQuery) []hive.Column {
	columns := []hive.Column{
		hive.Column{Name: "period_start", Type: "timestamp"},
		hive.Column{Name: "period_end", Type: "timestamp"},
	}
	for _, c := range genQuery.Spec.Columns {
		columns = append(columns, hive.Column{Name: c.Name, Type: c.Type})
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

func (c *Chargeback) newLogIdenti***REMOVED***er() logrus.Fields {
	return logrus.Fields{
		"logID": randomString(c.rand, logIdenti***REMOVED***erLength),
	}
}

package chargeback

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/sirupsen/logrus"
)

const logIdenti***REMOVED***erLength = 10

func init() {
	rand.Seed(time.Now().Unix())
}

var resourceNameReplacer = strings.NewReplacer("-", "_", ".", "_")

func dataSourceTableName(dataSourceName string) string {
	return fmt.Sprintf("datasource_%s", resourceNameReplacer.Replace(dataSourceName))
}

func reportTableName(reportName string) string {
	return fmt.Sprintf("report_%s", resourceNameReplacer.Replace(reportName))
}

func generationQueryViewName(queryName string) string {
	return fmt.Sprintf("view_%s", resourceNameReplacer.Replace(queryName))
}

func billingPeriodFormat(date time.Time) string {
	return date.Format(hive.HiveDateStringLayout)
}

func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func randomString(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func newLogIdenti***REMOVED***er() logrus.Fields {
	return logrus.Fields{
		"logID": randomString(logIdenti***REMOVED***erLength),
	}
}

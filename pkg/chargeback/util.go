package chargeback

import (
	"fmt"
	"strings"
	"time"
)

var resourceNameReplacer = strings.NewReplacer("-", "_", ".", "_")

func dataStoreTableName(dataStoreName string) string {
	return fmt.Sprintf("datastore_%s", resourceNameReplacer.Replace(dataStoreName))
}

func reportTableName(reportName string) string {
	return fmt.Sprintf("report_%s", resourceNameReplacer.Replace(reportName))
}

func generationQueryViewName(queryName string) string {
	return fmt.Sprintf("view_%s", resourceNameReplacer.Replace(queryName))
}

func truncateToSecond(t time.Time) time.Time {
	return t.Truncate(time.Second)
}

package chargeback

import (
	"fmt"
	"strings"
)

var resourceNameReplacer = strings.NewReplacer("-", "_", ".", "_")

func dataStoreTableName(dataStoreName string) string {
	return fmt.Sprintf("datastore_%s", resourceNameReplacer.Replace(dataStoreName))
}

func reportTableName(reportName string) string {
	return fmt.Sprintf("report_%s", resourceNameReplacer.Replace(reportName))
}

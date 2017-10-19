package chargeback

import (
	"database/sql/driver"
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

func argsString(args ...interface{}) string {
	var margs string
	for i, a := range args {
		var v interface{} = a
		if x, ok := v.(driver.Valuer); ok {
			y, err := x.Value()
			if err == nil {
				v = y
			}
		}
		switch v.(type) {
		case string, []byte:
			v = fmt.Sprintf("%q", v)
		default:
			v = fmt.Sprintf("%v", v)
		}
		margs += fmt.Sprintf("%d:%s", i+1, v)
		if i+1 < len(args) {
			margs += " "
		}
	}
	return margs
}

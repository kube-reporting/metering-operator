package reportingutil

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

var (
	resourceNameReplacer = strings.NewReplacer("-", "_", ".", "_")

	// AWSUsagePartitionDateStringLayout is the format used to partition
	// AWSUsage partition key
	AWSUsagePartitionDateStringLayout = "20060102"
)

func DataSourceTableName(namespace, dataSourceName string) string {
	return fmt.Sprintf("datasource_%s_%s", resourceNameReplacer.Replace(namespace), resourceNameReplacer.Replace(dataSourceName))
}

func ReportTableName(namespace, reportName string) string {
	return fmt.Sprintf("report_%s_%s", resourceNameReplacer.Replace(namespace), resourceNameReplacer.Replace(reportName))
}

func TableResourceNameFromKind(kind, namespace, name string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s-%s", kind, namespace, name))
}

func AWSBillingPeriodTimestamp(date time.Time) string {
	return date.Format(AWSUsagePartitionDateStringLayout)
}

func FullyQuali***REMOVED***edTableName(prestoTable *cbTypes.PrestoTable) string {
	return presto.FullyQuai***REMOVED***edTableName(prestoTable.Status.Catalog, prestoTable.Status.Schema, prestoTable.Status.TableName)
}

func IsValidSQLIdenti***REMOVED***er(id string) bool {
	if len(id) == 0 {
		return false
	}

	// First character must be a letter or underscore
	***REMOVED***rstChar := rune(id[0])
	if !unicode.IsLetter(***REMOVED***rstChar) && ***REMOVED***rstChar != '_' {
		return false
	}

	// Everything ***REMOVED*** character must be a letter, digit or underscore
	rest := id[1:]
	for _, r := range rest {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func TruncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func GenerateHiveColumns(query *cbTypes.ReportQuery) []hive.Column {
	var columns []hive.Column
	for _, col := range query.Spec.Columns {
		columns = append(columns, hive.Column{Name: col.Name, Type: col.Type})
	}
	return columns
}

func GeneratePrestoColumns(query *cbTypes.ReportQuery) []presto.Column {
	var columns []presto.Column
	for _, col := range query.Spec.Columns {
		columns = append(columns, presto.Column{Name: col.Name, Type: col.Type})
	}
	return columns
}

func HiveColumnsToPrestoColumns(columns []hive.Column) ([]presto.Column, error) {
	var err error
	newCols := make([]presto.Column, len(columns))
	for i, col := range columns {
		newCols[i], err = HiveColumnToPrestoColumn(col)
		if err != nil {
			return nil, err
		}
	}
	return newCols, nil
}

func PrestoColumnsToHiveColumns(columns []presto.Column) ([]hive.Column, error) {
	var err error
	newCols := make([]hive.Column, len(columns))
	for i, col := range columns {
		newCols[i], err = PrestoColumnToHiveColumn(col)
		if err != nil {
			return nil, err
		}
	}
	return newCols, nil
}

func SimpleHiveColumnTypeToPrestoColumnType(colType string) string {
	colType = strings.ToUpper(colType)
	switch colType {
	case "TINYINT", "SMALLINT", "INT", "INTEGER", "BIGINT",
		"FLOAT", "DOUBLE", "BOOLEAN",
		"VARCHAR", "CHAR",
		"DATE", "TIME", "TIMESTAMP", "BINARY":
		return colType
	case "STRING":
		return "VARCHAR"
	}
	return ""
}

func HiveColumnToPrestoColumn(column hive.Column) (presto.Column, error) {
	colType := SimpleHiveColumnTypeToPrestoColumnType(column.Type)
	if colType != "" {
		return presto.Column{
			Name: column.Name,
			Type: colType,
		}, nil
	} ***REMOVED*** {
		colType = strings.ToUpper(column.Type)
		switch {
		case strings.Contains(colType, "MAP"):
			// does not support maps with arrays inside them
			if strings.Contains(colType, "ARRAY") {
				return presto.Column{}, fmt.Errorf("cannot convert map containing array into map for Presto, column: %q, type: %q", column.Name, column.Type)
			}
			beginMapIndex := strings.Index(colType, "<")
			endMapIndex := strings.Index(colType, ">")
			if beginMapIndex == -1 || endMapIndex == -1 {
				return presto.Column{}, fmt.Errorf("unable to ***REMOVED***nd matching <, > pair for column %q, type: %q", column.Name, column.Type)
			}
			if beginMapIndex+1 >= len(colType) {
				return presto.Column{}, fmt.Errorf("invalid map de***REMOVED***nition in column type, column %q, type: %q", column.Name, column.Type)
			}
			mapComponents := colType[beginMapIndex+1 : endMapIndex]
			mapComponentsSplit := strings.SplitN(mapComponents, ",", 2)
			if len(mapComponentsSplit) != 2 {
				return presto.Column{}, fmt.Errorf("invalid map de***REMOVED***nition in column type, column %q, type: %q", column.Name, column.Type)
			}
			keyType := strings.TrimSpace(mapComponentsSplit[0])
			valueType := strings.TrimSpace(mapComponentsSplit[1])

			prestoKeyType := SimpleHiveColumnTypeToPrestoColumnType(keyType)
			prestoValueType := SimpleHiveColumnTypeToPrestoColumnType(valueType)
			if prestoKeyType == "" {
				return presto.Column{}, fmt.Errorf("invalid presto map key type: %q", keyType)
			}
			if prestoValueType == "" {
				return presto.Column{}, fmt.Errorf("invalid presto map value type: %q", valueType)
			}
			mapColType := fmt.Sprintf("map(%s,%s)", prestoKeyType, prestoValueType)

			return presto.Column{
				Name: column.Name,
				Type: mapColType,
			}, nil

		case strings.Contains(colType, "ARRAY"):
			// currently unsupported
		}
	}
	return presto.Column{}, fmt.Errorf("unsupported hive type: %q", column.Type)
}

func SimplePrestoColumnTypeToHiveColumnType(colType string) string {
	colType = strings.ToUpper(colType)
	switch colType {
	case "TINYINT", "SMALLINT", "INT", "INTEGER", "BIGINT",
		"FLOAT", "DOUBLE", "BOOLEAN",
		"CHAR",
		"DATE", "TIME", "TIMESTAMP", "BINARY":
		return colType
	case "REAL":
		return "FLOAT"
	case "VARCHAR":
		return "STRING"
	}
	return ""
}

func PrestoColumnToHiveColumn(column presto.Column) (hive.Column, error) {
	colType := SimplePrestoColumnTypeToHiveColumnType(column.Type)
	if colType != "" {
		return hive.Column{
			Name: column.Name,
			Type: colType,
		}, nil
	} ***REMOVED*** {
		colType = strings.ToUpper(column.Type)
		switch {
		case strings.Contains(colType, "MAP"):
			// does not support maps with arrays inside them
			if strings.Contains(colType, "ARRAY") {
				return hive.Column{}, fmt.Errorf("cannot convert map containing array into map for Presto, column: %q, type: %q", column.Name, column.Type)
			}
			beginMapIndex := strings.Index(colType, "(")
			endMapIndex := strings.Index(colType, ")")
			if beginMapIndex == -1 || endMapIndex == -1 {
				return hive.Column{}, fmt.Errorf("unable to ***REMOVED***nd matching (, ) pair for column %q, type: %q", column.Name, column.Type)
			}
			if beginMapIndex+1 >= len(colType) {
				return hive.Column{}, fmt.Errorf("invalid map de***REMOVED***nition in column type, column %q, type: %q", column.Name, column.Type)
			}
			mapComponents := colType[beginMapIndex+1 : endMapIndex]
			mapComponentsSplit := strings.SplitN(mapComponents, ",", 2)
			if len(mapComponentsSplit) != 2 {
				return hive.Column{}, fmt.Errorf("invalid map de***REMOVED***nition in column type, column %q, type: %q", column.Name, column.Type)
			}
			keyType := strings.TrimSpace(mapComponentsSplit[0])
			valueType := strings.TrimSpace(mapComponentsSplit[1])

			hiveKeyType := SimplePrestoColumnTypeToHiveColumnType(keyType)
			hiveValueType := SimplePrestoColumnTypeToHiveColumnType(valueType)
			if hiveKeyType == "" {
				return hive.Column{}, fmt.Errorf("invalid hive map key type: %q", keyType)
			}
			if hiveValueType == "" {
				return hive.Column{}, fmt.Errorf("invalid hive map value type: %q", valueType)
			}
			mapColType := fmt.Sprintf("MAP<%s,%s>", hiveKeyType, hiveValueType)

			return hive.Column{
				Name: column.Name,
				Type: mapColType,
			}, nil

		case strings.Contains(colType, "ARRAY"):
			// currently unsupported
		}
	}
	return hive.Column{}, fmt.Errorf("unsupported hive type: %q", column.Type)
}

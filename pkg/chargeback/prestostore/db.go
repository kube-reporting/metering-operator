package prestostore

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

const (
	// prestoQueryCap is the maximum payload size a single SQL statement can contain
	// before Presto will error due to the payload being too large.
	prestoQueryCap = 1000000
	// prestoTimestampFormat is the format Presto expects timetamps to look
	// like. For use with time.Format()
	prestoTimestampFormat = "2006-01-02 15:04:05.000"
)

// StorePrometheusMetrics handles storing Prometheus metrics into the speci***REMOVED***ed
// Presto table.
//
// Any Queryer is accepted, but this function expects a Presto connection.
func StorePrometheusMetrics(ctx context.Context, queryer db.Queryer, tableName string, metrics []*PrometheusMetric) error {
	var queryValues []string

	for _, metric := range metrics {
		metricValue := generatePrometheusMetricSQLValues(metric)
		queryValues = append(queryValues, metricValue)
	}
	// capacity prestoQueryCap, length 0
	queryBuf := bytes.NewBuffer(make([]byte, 0, prestoQueryCap))

	insertStatementLength := len(presto.FormatInsertQuery(tableName, ""))
	// calculate the queryCap with the "INSERT INTO $table_name" portion
	// accounted for
	queryCap := prestoQueryCap - insertStatementLength

	for _, value := range queryValues {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		// If the buffer is empty, we add VALUES to it, and everything the
		// follows will be a single row to insert
		if queryBuf.Len() == 0 {
			queryBuf.WriteString("VALUES ")
		} ***REMOVED*** {
			// if the buffer isn't empty, then before we add more rows to the
			// insert query, add a comma to separate them.
			queryBuf.WriteString(",")
		}

		// There's a character limit of prestoQueryCap on insert
		// queries, so let's chunk them at that limit.
		bytesToWrite := len(value)
		newBufferSize := (bytesToWrite + queryBuf.Len())

		// if writing the current value to the buffer would exceed the
		// prestoQueryCap, preform the insert query, and reset the buffer
		if newBufferSize > queryCap {
			err := presto.ExecuteInsertQuery(queryer, tableName, queryBuf.String())
			if err != nil {
				return fmt.Errorf("failed to store metrics into presto: %v", err)
			}
			queryBuf.Reset()
		} ***REMOVED*** {
			queryBuf.WriteString(value)
		}
	}
	// if the buffer has unwritten values, perform the ***REMOVED***nal insert
	if queryBuf.Len() != 0 {
		err := presto.ExecuteInsertQuery(queryer, tableName, queryBuf.String())
		if err != nil {
			return fmt.Errorf("failed to store metrics into presto: %v", err)
		}
	}
	return nil
}

// generatePrometheusMetricSQLValues turns a PrometheusMetric into a SQL literal
// suited for INSERT statements. To insert maps, we crete an array of keys and
// values as recommended by Presto documentation.
//
// The schema is as follows:
// column "amount" type: "double"
// column "timestamp" type: "timestamp"
// column "timePrecision" type: "double"
// column "labels" type: "map<string, string>"
func generatePrometheusMetricSQLValues(metric *PrometheusMetric) string {
	var keys []string
	var vals []string
	for k, v := range metric.Labels {
		keys = append(keys, "'"+k+"'")
		vals = append(vals, "'"+v+"'")
	}
	keyString := "ARRAY[" + strings.Join(keys, ",") + "]"
	valString := "ARRAY[" + strings.Join(vals, ",") + "]"
	return fmt.Sprintf("(%f,timestamp '%s',%f,map(%s,%s))",
		metric.Amount, metric.Timestamp.Format(prestoTimestampFormat), metric.StepSize.Seconds(), keyString, valString)
}

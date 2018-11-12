package prestostore

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

const (
	// prestoQueryCap is the maximum payload size a single SQL statement can contain
	// before Presto will error due to the payload being too large.
	prestoQueryCap = 1000000
)

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			// capacity prestoQueryCap, length 0
			return bytes.NewBuffer(make([]byte, 0, prestoQueryCap))
		},
	}

	promsumColumns = []presto.Column{
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map(varchar, varchar)"},
	}
)

type PrometheusMetricsStorer interface {
	StorePrometheusMetrics(ctx context.Context, tableName string, metrics []*PrometheusMetric) error
}

type PrometheusMetricsGetter interface {
	GetPrometheusMetrics(tableName string, start, end time.Time) ([]*PrometheusMetric, error)
}

type PrometheusMetricTimestampTracker interface {
	GetLastTimestampForTable(tableName string) (*time.Time, error)
}

type PrometheusMetricsRepo interface {
	PrometheusMetricsGetter
	PrometheusMetricsStorer
	PrometheusMetricTimestampTracker
}

type prometheusMetricRepo struct {
	queryer db.Queryer
}

func NewPrometheusMetricsRepo(queryer db.Queryer) *prometheusMetricRepo {
	return &prometheusMetricRepo{
		queryer: queryer,
	}
}

func (r *prometheusMetricRepo) StorePrometheusMetrics(ctx context.Context, tableName string, metrics []*PrometheusMetric) error {
	return StorePrometheusMetrics(ctx, r.queryer, tableName, metrics)
}

func (r *prometheusMetricRepo) GetPrometheusMetrics(tableName string, start, end time.Time) ([]*PrometheusMetric, error) {
	return GetPrometheusMetrics(r.queryer, tableName, start, end)
}

func (r *prometheusMetricRepo) GetLastTimestampForTable(tableName string) (*time.Time, error) {
	// Get the most recent timestamp in the table for this query
	getLastTimestampQuery := fmt.Sprintf(`
				SELECT "timestamp"
				FROM %s
				ORDER BY "timestamp" DESC
				LIMIT 1`, tableName)

	results, err := presto.ExecuteSelect(r.queryer, getLastTimestampQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting last timestamp for table %s, maybe table doesn't exist yet? %v", tableName, err)
	}

	if len(results) != 0 {
		ts := results[0]["timestamp"].(time.Time)
		return &ts, nil
	}
	return nil, nil
}

// PrometheusMetric is a receipt of a usage determined by a query within a specific time range.
type PrometheusMetric struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}

// StorePrometheusMetrics handles storing Prometheus metrics into the specified
// Presto table.
func StorePrometheusMetrics(ctx context.Context, queryer db.Queryer, tableName string, metrics []*PrometheusMetric) error {
	queryBuf := bufPool.Get().(*bytes.Buffer)
	queryBuf.Reset()
	defer bufPool.Put(queryBuf)

	insertStatementLength := len(presto.FormatInsertQuery(tableName, ""))
	// calculate the queryCap with the "INSERT INTO $table_name" portion
	// accounted for
	queryCap := prestoQueryCap - insertStatementLength

	for _, metric := range metrics {
		metricValue := generatePrometheusMetricSQLValues(metric)

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
		} else {
			// if the buffer isn't empty, then before we add more rows to the
			// insert query, add a comma to separate them.
			queryBuf.WriteString(",")
		}

		// There's a character limit of prestoQueryCap on insert
		// queries, so let's chunk them at that limit.
		bytesToWrite := len(metricValue)
		newBufferSize := (bytesToWrite + queryBuf.Len())

		// if writing the current metricValue to the buffer would exceed the
		// prestoQueryCap, preform the insert query, and reset the buffer
		if newBufferSize > queryCap {
			err := presto.InsertInto(queryer, tableName, queryBuf.String())
			if err != nil {
				return fmt.Errorf("failed to store metrics into presto: %v", err)
			}
			queryBuf.Reset()
		} else {
			queryBuf.WriteString(metricValue)
		}
	}
	// if the buffer has unwritten values, perform the final insert
	if queryBuf.Len() != 0 {
		err := presto.InsertInto(queryer, tableName, queryBuf.String())
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
// the following columns are partition columns:
// column "dt" type: "string"
func generatePrometheusMetricSQLValues(metric *PrometheusMetric) string {
	var keys []string
	var vals []string
	for k, v := range metric.Labels {
		keys = append(keys, "'"+k+"'")
		vals = append(vals, "'"+v+"'")
	}
	keyString := "ARRAY[" + strings.Join(keys, ",") + "]"
	valString := "ARRAY[" + strings.Join(vals, ",") + "]"
	dt := PrometheusMetricTimestampPartition(metric.Timestamp)
	return fmt.Sprintf("(%f,timestamp '%s',%f,map(%s,%s),'%s')",
		metric.Amount, metric.Timestamp.Format(presto.TimestampFormat), metric.StepSize.Seconds(), keyString, valString, dt,
	)
}

const PrometheusMetricTimestampPartitionFormat = "2006-01-02"

func PrometheusMetricTimestampPartition(t time.Time) string {
	return t.UTC().Format(PrometheusMetricTimestampPartitionFormat)
}

func GetPrometheusMetrics(queryer db.Queryer, tableName string, start, end time.Time) ([]*PrometheusMetric, error) {
	whereClause := ""
	if !start.IsZero() {
		whereClause += fmt.Sprintf(`WHERE "timestamp" >= timestamp '%s' `, start.Format(presto.TimestampFormat))
	}
	if !end.IsZero() {
		if !start.IsZero() {
			whereClause += " AND "
		} else {
			whereClause += " WHERE "
		}
		whereClause += fmt.Sprintf(`"timestamp" <= timestamp '%s'`, end.Format(presto.TimestampFormat))
	}

	rows, err := presto.GetRows(queryer, tableName, promsumColumns)
	if err != nil {
		return nil, err
	}

	results := make([]*PrometheusMetric, len(rows))
	for i, row := range rows {
		rowLabels := row["labels"].(map[string]interface{})
		rowAmount := row["amount"].(float64)
		rowTimePrecision := row["timeprecision"].(float64)
		rowTimestamp := row["timestamp"].(time.Time)

		labels := make(map[string]string)
		for key, value := range rowLabels {
			var ok bool
			labels[key], ok = value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid label %s, valueType: %T, value: %+v", key, value, value)
			}
		}
		metric := &PrometheusMetric{
			Labels:    labels,
			Amount:    rowAmount,
			StepSize:  time.Duration(rowTimePrecision) * time.Second,
			Timestamp: rowTimestamp,
		}
		results[i] = metric
	}
	return results, nil
}

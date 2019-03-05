package prestostore

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

const (
	// defaultPrestoQueryCap is the default maximum payload size a single SQL
	// statement can contain before Presto will error due to the payload being
	// too large.
	defaultPrestoQueryCap = 1000000
)

var (
	defaultQueryBufferPool = NewBufferPool(defaultPrestoQueryCap)

	PromsumHiveTableColumns = []hive.Column{
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map<string, string>"},
	}
	PromsumHivePartitionColumns = []hive.Column{
		{Name: "dt", Type: "string"},
	}

	// Initialized by init()
	PromsumPrestoTableColumn, PromsumPrestoPartitionColumns, PromsumPrestoAllColumns []presto.Column
)

func init() {
	var err error
	PromsumPrestoTableColumn, err = reportingutil.HiveColumnsToPrestoColumns(PromsumHiveTableColumns)
	if err != nil {
		panic(err)
	}
	PromsumPrestoPartitionColumns, err = reportingutil.HiveColumnsToPrestoColumns(PromsumHivePartitionColumns)
	if err != nil {
		panic(err)
	}
	PromsumPrestoAllColumns = append(PromsumPrestoTableColumn, PromsumPrestoPartitionColumns...)
}

func NewBufferPool(capacity int) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, capacity))
		},
	}

}

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
	queryer         db.Queryer
	queryBufferPool sync.Pool
}

func NewPrometheusMetricsRepo(queryer db.Queryer, queryBufferPool *sync.Pool) *prometheusMetricRepo {
	if queryBufferPool == nil {
		queryBufferPool = &defaultQueryBufferPool
	}
	return &prometheusMetricRepo{
		queryer:         queryer,
		queryBufferPool: *queryBufferPool,
	}
}

func (r *prometheusMetricRepo) StorePrometheusMetrics(ctx context.Context, tableName string, metrics []*PrometheusMetric) error {
	queryBuf := r.queryBufferPool.Get().(*bytes.Buffer)
	queryBuf.Reset()
	defer r.queryBufferPool.Put(queryBuf)
	return StorePrometheusMetricsWithBuffer(queryBuf, ctx, r.queryer, tableName, metrics)
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
	Dt        string            `json:"dt"`
}

// storePrometheusMetricsWithBuffer handles storing Prometheus metrics into the
// specified Presto table.
func StorePrometheusMetricsWithBuffer(queryBuf *bytes.Buffer, ctx context.Context, queryer db.Queryer, tableName string, metrics []*PrometheusMetric) error {
	bufferCapacity := queryBuf.Cap()

	insertStatementLength := len(presto.FormatInsertQuery(tableName, ""))
	// calculate the queryCap with the "INSERT INTO $table_name" portion
	// accounted for
	queryCap := bufferCapacity - insertStatementLength
	// account for "," and "VALUES " string length when writing to buffer
	commaStr := ","
	valuesStmtStr := "VALUES "

	metricsInBuffer := false
	numMetrics := len(metrics)

	for i, metric := range metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		metricSQLStr := generatePrometheusMetricSQLValues(metric)

		// lastMetric means we need to insert after writing the metric to the
		// buffer
		lastMetric := i == (numMetrics - 1)

		if metricsInBuffer {
			// if writing the current metricSQLStr to the buffer would exceed the
			// bufferCapacity, perform the insert query, and reset the buffer
			// to flush it
			bytesToWrite := len(commaStr + metricSQLStr)
			if (bytesToWrite + queryBuf.Len()) > queryCap {
				err := presto.InsertInto(queryer, tableName, queryBuf.String())
				if err != nil {
					return fmt.Errorf("failed to store metrics into presto: %v", err)
				}
				queryBuf.Reset()

				// we just inserted the contents of the buffer, so reset
				// metricsInBuffer and prepend VALUES
				metricsInBuffer = false
			}
		}

		var toWrite string
		if !metricsInBuffer {
			// no metrics in buffer means we need to prepend "VALUES " before
			// we write metricSQL
			toWrite = valuesStmtStr + metricSQLStr
		} else {
			// existing metrics in buffer means we need to prepend "," before
			// we write metricSQL since that separates each record
			toWrite = commaStr + metricSQLStr
		}

		bytesToWrite := len(toWrite)
		if (bytesToWrite + queryBuf.Len()) > queryCap {
			return fmt.Errorf("writing %q would exceed buffer size, please adjust buffer size: bufferCapacityBytes: %d, queryCapacityBytes: %d, currentBufferSize: %d bytesToWrite: %d", toWrite, bufferCapacity, queryCap, queryBuf.Len(), bytesToWrite)
		}

		_, err := queryBuf.WriteString(toWrite)
		if err != nil {
			return fmt.Errorf(`error writing %q string to buffer: %v`, toWrite, err)
		}
		metricsInBuffer = true

		// this is the last metric in the loop, insert the contents of the
		// buffer
		if lastMetric {
			err := presto.InsertInto(queryer, tableName, queryBuf.String())
			if err != nil {
				return fmt.Errorf("failed to store metrics into presto: %v", err)
			}
			queryBuf.Reset()
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

	rows, err := presto.GetRows(queryer, tableName, PromsumPrestoAllColumns)
	if err != nil {
		return nil, err
	}

	results := make([]*PrometheusMetric, len(rows))
	for i, row := range rows {
		rowLabels := row["labels"].(map[string]interface{})
		rowAmount := row["amount"].(float64)
		rowTimePrecision := row["timePrecision"].(float64)
		rowTimestamp := row["timestamp"].(time.Time)
		dt := row["dt"].(string)

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
			Dt:        dt,
		}
		results[i] = metric
	}
	return results, nil
}

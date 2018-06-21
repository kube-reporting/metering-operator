package promexporter

import (
	"context"
	"fmt"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/promcollector"
)

const (
	// cap the maximum c.cfg.ChunkSize
	maxChunkDuration = 24 * time.Hour
)

// PrestoExporter exports Prometheus metrics into Presto tables
type PrestoExporter struct {
	logger        logrus.FieldLogger
	prestoQueryer db.Queryer
	promConn      prom.API
	clock         clock.Clock
	cfg           Config

	lastTimestamp *time.Time
}

type Config struct {
	PrometheusQuery       string
	PrestoTableName       string
	ChunkSize             time.Duration
	StepSize              time.Duration
	MaxTimeRanges         int64
	AllowIncompleteChunks bool
}

func NewPrestoExporter(logger logrus.FieldLogger, promConn prom.API, prestoQueryer db.Queryer, clock clock.Clock, cfg Config) *PrestoExporter {
	return &PrestoExporter{
		logger: logger.WithFields(logrus.Fields{
			"component": "PrestoExporter",
			"tableName": cfg.PrestoTableName,
		}),
		prestoQueryer: prestoQueryer,
		promConn:      promConn,
		clock:         clock,
		cfg:           cfg,
	}
}

// Export triggers the PrestoExporter to query Prometheus and store the results
// in Presto. It will block until collection is finished.

// It queries Prometheus using promQuery, storing the results in tableName. The
// exporter will track the last time series it retrieved and will pick up from
// where it left off if paused or stopped. FOr more details, see
// pkg/promcollector.
func (c *PrestoExporter) Export(ctx context.Context) error {
	logger := c.logger
	logger.Infof("PrestoExporter started")

	endTime := c.clock.Now()

	if c.lastTimestamp != nil {
		c.logger.Debugf("got lastTimestamp for table %s: %s", c.cfg.PrestoTableName, c.lastTimestamp.String())
	} else {
		// Looks like we haven't populated any data in this table yet.
		// Let's backfill our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == endTime, so it won't be queried, so this gets the chunk
		// before the latest
		backfillUntil := endTime.Add(-2 * c.cfg.ChunkSize)
		c.lastTimestamp = &backfillUntil
		c.logger.Debugf("no data in data store %s yet", c.cfg.PrestoTableName)
	}

	preProcessingHandler := func(_ context.Context, timeRanges []prom.Range) error {
		if len(timeRanges) == 0 {
			logger.Info("no time ranges to query yet for table %s", c.cfg.PrestoTableName)
		} else {
			begin := timeRanges[0].Start
			end := timeRanges[len(timeRanges)-1].End
			logger.Infof("querying for data between %s and %s (chunks: %d)", begin, end, len(timeRanges))
		}
		return nil
	}

	postQueryHandler := func(ctx context.Context, timeRange prom.Range, matrix model.Matrix) error {
		records := promMatrixToRecords(timeRange, matrix)
		err := storePrometheusRecords(ctx, c.prestoQueryer, c.cfg.PrestoTableName, records)
		if err != nil {
			return fmt.Errorf("failed to store Prometheus metrics into table %s for the range %v to %v: %v",
				c.cfg.PrestoTableName, timeRange.Start, timeRange.End, err)
		}
		return nil
	}

	collector := promcollector.New(c.promConn, c.cfg.PrometheusQuery, preProcessingHandler, nil, postQueryHandler, nil)

	// if c.lastTimestamp is null then it's because we errored sometime
	// last time we collected and need to re-query Presto to figure out
	// the last timestamp
	if c.lastTimestamp == nil {
		var err error
		c.lastTimestamp, err = getLastTimestampForTable(c.prestoQueryer, c.cfg.PrestoTableName)
		if err != nil {
			logger.WithError(err).Errorf("unable to get last timestamp for table %s", c.cfg.PrestoTableName)
			return nil
		}
	}
	// We don't want to duplicate the c.lastTimestamp record so add
	// the step size so that we start at the next interval no longer in
	// our range.
	startTime := c.lastTimestamp.Add(c.cfg.StepSize)

	// If the c.lastTimestamp is too far back, we should limit this run to
	// maxChunkDuration so that if we're stopped for an extended amount of time,
	// this function won't return a slice with too many time ranges.
	totalChunkDuration := c.lastTimestamp.Sub(endTime)
	if totalChunkDuration >= maxChunkDuration {
		endTime = c.lastTimestamp.Add(maxChunkDuration)
	}

	loggerWithFields := logger.WithFields(logrus.Fields{
		"startTime": startTime,
		"endTime":   endTime,
	})

	timeRangesCollected, err := collector.Collect(ctx, startTime, endTime, c.cfg.ChunkSize, c.cfg.StepSize, c.cfg.MaxTimeRanges, c.cfg.AllowIncompleteChunks)
	if err != nil {
		loggerWithFields.WithError(err).Error("error collecting metrics")
		// at this point we cannot be sure what is in Presto and what
		// isn't, so reset our c.lastTimestamp
		c.lastTimestamp = nil
	}

	if len(timeRangesCollected) == 0 {
		loggerWithFields.Infof("no data collected for table %s", c.cfg.PrestoTableName)
		return nil
	}

	// update our c.lastTimestamp
	lastTS := timeRangesCollected[len(timeRangesCollected)-1].End
	c.lastTimestamp = &lastTS

	logger.Infof("PrestoExporter finished")
	return nil
}

func Collect(ctx context.Context, promConn prom.API, queryer db.Queryer, query string, tableName string, startTime, endTime time.Time, chunkSize, stepSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool) error {
	postQueryHandler := func(ctx context.Context, timeRange prom.Range, matrix model.Matrix) error {
		records := promMatrixToRecords(timeRange, matrix)
		return storePrometheusRecords(ctx, queryer, tableName, records)
	}
	collector := promcollector.New(promConn, query, nil, nil, postQueryHandler, nil)
	_, err := collector.Collect(ctx, startTime, endTime, chunkSize, stepSize, maxTimeRanges, allowIncompleteChunks)
	return err
}

func getLastTimestampForTable(queryer db.Queryer, tableName string) (*time.Time, error) {
	// Get the most recent timestamp in the table for this query
	getLastTimestampQuery := fmt.Sprintf(`
				SELECT "timestamp"
				FROM %s
				ORDER BY "timestamp" DESC
				LIMIT 1`, tableName)

	results, err := presto.ExecuteSelect(queryer, getLastTimestampQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting last timestamp for table %s, maybe table doesn't exist yet? %v", tableName, err)
	}

	if len(results) != 0 {
		ts := results[0]["timestamp"].(time.Time)
		return &ts, nil
	}
	return nil, nil
}

func promMatrixToRecords(timeRange prom.Range, matrix model.Matrix) []*Record {
	var records []*Record
	// iterate over segments of contiguous billing records
	for _, sampleStream := range matrix {
		for _, value := range sampleStream.Values {
			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			record := &Record{
				Labels:    labels,
				Amount:    float64(value.Value),
				StepSize:  timeRange.Step,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records
}

// Record is a receipt of a usage determined by a query within a specific time range.
type Record struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}

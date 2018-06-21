package promexporter

import (
	"context"
	"fmt"
	"sync"
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
	logger          logrus.FieldLogger
	promConn        prom.API
	prestoQueryer   db.Queryer
	collectHandlers promcollector.CollectHandlers
	clock           clock.Clock
	cfg             Con***REMOVED***g

	// exportLock ensures only one export is running at a time, protecting the
	// lastTimestamp and records ***REMOVED***elds
	exportLock sync.Mutex

	//lastTimestamp is the lastTimestamp stored for this PrestoExporter
	lastTimestamp *time.Time
	// records contains the records we stored after an export. It is a pointer
	// to a slice so we can change the contents from the postQueryHandler of
	// our promcollector.Collector. After the export is ***REMOVED***nished it is expected that records is cleared.
	records *[]*Record
}

type Con***REMOVED***g struct {
	PrometheusQuery       string
	PrestoTableName       string
	ChunkSize             time.Duration
	StepSize              time.Duration
	MaxTimeRanges         int64
	AllowIncompleteChunks bool
}

func NewPrestoExporter(logger logrus.FieldLogger, promConn prom.API, prestoQueryer db.Queryer, clock clock.Clock, cfg Con***REMOVED***g) *PrestoExporter {
	preProcessingHandler := func(_ context.Context, timeRanges []prom.Range) error {
		if len(timeRanges) == 0 {
			logger.Info("no time ranges to query yet for table %s", cfg.PrestoTableName)
		} ***REMOVED*** {
			begin := timeRanges[0].Start
			end := timeRanges[len(timeRanges)-1].End
			logger.Infof("querying for data between %s and %s (chunks: %d)", begin, end, len(timeRanges))
		}
		return nil
	}

	var recordsPtr *[]*Record
	postQueryHandler := func(ctx context.Context, timeRange prom.Range, matrix model.Matrix) error {
		records := promMatrixToRecords(timeRange, matrix)
		err := StorePrometheusRecords(ctx, prestoQueryer, cfg.PrestoTableName, records)
		if err != nil {
			return fmt.Errorf("failed to store Prometheus metrics into table %s for the range %v to %v: %v",
				cfg.PrestoTableName, timeRange.Start, timeRange.End, err)
		}
		recordsPtr = &records

		return nil
	}

	collectHandlers := promcollector.CollectHandlers{
		PreProcessingHandler: preProcessingHandler,
		PostQueryHandler:     postQueryHandler,
	}

	return &PrestoExporter{
		logger: logger.WithFields(logrus.Fields{
			"component": "PrestoExporter",
			"tableName": cfg.PrestoTableName,
		}),
		promConn:        promConn,
		prestoQueryer:   prestoQueryer,
		collectHandlers: collectHandlers,
		clock:           clock,
		cfg:             cfg,
		records:         recordsPtr,
	}
}

// ExportFromLastTimestamp executes a Presto query from the last time range it
// queried and stores the results in a Presto table.

// The exporter will track the last time series it retrieved and will query
// the next time range starting from where it left off if paused or stopped.
// For more details on how querying Prometheus is done, see the package
// pkg/promcollector.
func (c *PrestoExporter) ExportFromLastTimestamp(ctx context.Context) ([]*Record, error) {
	c.exportLock.Lock()
	defer c.exportLock.Unlock()
	logger := c.logger
	logger.Infof("PrestoExporter ExportFromLastTimestamp started")

	endTime := c.clock.Now()

	if c.lastTimestamp != nil {
		c.logger.Debugf("got lastTimestamp for table %s: %s", c.cfg.PrestoTableName, c.lastTimestamp.String())
	} ***REMOVED*** {
		// Looks like we haven't populated any data in this table yet.
		// Let's back***REMOVED***ll our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == endTime, so it won't be queried, so this gets the chunk
		// before the latest
		back***REMOVED***llUntil := endTime.Add(-2 * c.cfg.ChunkSize)
		c.lastTimestamp = &back***REMOVED***llUntil
		c.logger.Debugf("no data in data store %s yet", c.cfg.PrestoTableName)
	}

	// if c.lastTimestamp is null then it's because we errored sometime
	// last time we collected and need to re-query Presto to ***REMOVED***gure out
	// the last timestamp
	if c.lastTimestamp == nil {
		var err error
		c.lastTimestamp, err = getLastTimestampForTable(c.prestoQueryer, c.cfg.PrestoTableName)
		if err != nil {
			logger.WithError(err).Errorf("unable to get last timestamp for table %s", c.cfg.PrestoTableName)
			return nil, err
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

	timeRangesCollected, err := c.export(ctx, startTime, endTime)
	if err != nil {
		loggerWithFields.WithError(err).Error("error collecting metrics")
	}

	if len(timeRangesCollected) == 0 {
		loggerWithFields.Infof("no data collected for table %s", c.cfg.PrestoTableName)
		return timeRangesCollected, nil
	}

	logger.Infof("PrestoExporter ExportFromLastTimestamp ***REMOVED***nished")
	return timeRangesCollected, nil

}

func (c *PrestoExporter) Export(ctx context.Context, startTime, endTime time.Time) ([]*Record, error) {
	c.exportLock.Lock()
	defer c.exportLock.Unlock()
	logger := c.logger
	logger.Infof("PrestoExporter Export started")
	loggerWithFields := logger.WithFields(logrus.Fields{
		"startTime": startTime,
		"endTime":   endTime,
	})

	timeRangesCollected, err := c.export(ctx, startTime, endTime)
	if err != nil {
		loggerWithFields.WithError(err).Error("error collecting metrics")
	}

	if len(timeRangesCollected) == 0 {
		loggerWithFields.Infof("no data collected for table %s", c.cfg.PrestoTableName)
		return timeRangesCollected, nil
	}
	logger.Infof("PrestoExporter Export ***REMOVED***nished")
	return timeRangesCollected, nil
}

func (c *PrestoExporter) export(ctx context.Context, startTime, endTime time.Time) ([]*Record, error) {
	_, err := promcollector.Collect(ctx, c.promConn, c.cfg.PrometheusQuery, startTime, endTime, c.cfg.ChunkSize, c.cfg.StepSize, c.cfg.MaxTimeRanges, c.cfg.AllowIncompleteChunks, c.collectHandlers)
	if err != nil {
		// at this point we cannot be sure what is in Presto and what
		// isn't, so reset our c.lastTimestamp
		c.lastTimestamp = nil
	}

	var records []*Record
	if c.records != nil {
		// keep a reference to our slice and
		// reset c.records everytime we export it starts unset
		records = *c.records
		c.records = nil
	}

	// if we got records, update our cached lastTimestamp
	if err == nil && len(records) != 0 {
		lastTS := records[len(records)-1].Timestamp
		c.lastTimestamp = &lastTS
	}

	return records, err
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

// Record is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type Record struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}

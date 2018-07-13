package prestostore

import (
	"context"
	"fmt"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"

	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/promquery"
)

const (
	// cap the maximum c.cfg.ChunkSize
	maxChunkDuration = 24 * time.Hour
)

// PrometheusImporter imports Prometheus metrics into Presto tables
type PrometheusImporter struct {
	logger        logrus.FieldLogger
	promConn      prom.API
	prestoQueryer presto.ExecQueryer
	clock         clock.Clock
	cfg           Con***REMOVED***g

	// importLock ensures only one import is running at a time, protecting the
	// lastTimestamp and metrics ***REMOVED***elds
	importLock sync.Mutex

	//lastTimestamp is the lastTimestamp stored for this PrometheusImporter
	lastTimestamp *time.Time
	metricsCount  int
}

type Con***REMOVED***g struct {
	PrometheusQuery       string
	PrestoTableName       string
	ChunkSize             time.Duration
	StepSize              time.Duration
	MaxTimeRanges         int64
	MaxQueryRangeDuration time.Duration
}

func NewPrometheusImporter(logger logrus.FieldLogger, promConn prom.API, prestoQueryer presto.ExecQueryer, clock clock.Clock, cfg Con***REMOVED***g) *PrometheusImporter {
	logger = logger.WithFields(logrus.Fields{
		"component": "PrometheusImporter",
		"tableName": cfg.PrestoTableName,
		"chunkSize": cfg.ChunkSize,
		"stepSize":  cfg.StepSize,
	})

	return &PrometheusImporter{
		logger:        logger,
		promConn:      promConn,
		prestoQueryer: prestoQueryer,
		clock:         clock,
		cfg:           cfg,
	}
}

func (c *PrometheusImporter) preProcessingHandler(_ context.Context, timeRanges []prom.Range) error {
	// reset counter before we begin processing
	c.metricsCount = 0

	if len(timeRanges) == 0 {
		c.logger.Infof("no time ranges to query yet for table %s", c.cfg.PrestoTableName)
	} ***REMOVED*** {
		begin := timeRanges[0].Start.UTC()
		end := timeRanges[len(timeRanges)-1].End.UTC()
		logger := c.logger.WithFields(logrus.Fields{
			"rangeBegin": begin,
			"rangeEnd":   end,
		})
		logger.Debugf("querying for data between %s and %s (chunks: %d)", begin, end, len(timeRanges))
	}
	return nil
}

func (c *PrometheusImporter) preQueryHandler(ctx context.Context, timeRange prom.Range) error {
	queryBegin := timeRange.Start.UTC()
	queryEnd := timeRange.End.UTC()
	logger := c.logger.WithFields(logrus.Fields{
		"queryBegin": queryBegin,
		"queryEnd":   queryEnd,
	})

	logger.Debugf("querying Prometheus using range %s to %s", timeRange.Start, timeRange.End)
	return nil
}

func (c *PrometheusImporter) postQueryHandler(ctx context.Context, timeRange prom.Range, matrix model.Matrix) error {
	queryBegin := timeRange.Start.UTC()
	queryEnd := timeRange.End.UTC()

	metrics := promMatrixToPrometheusMetrics(timeRange, matrix)
	if len(metrics) != 0 {
		metricsBegin := metrics[0].Timestamp
		metricsEnd := metrics[len(metrics)-1].Timestamp
		metricLogger := c.logger.WithFields(logrus.Fields{
			"metricsBegin": metricsBegin,
			"metricsEnd":   metricsEnd,
		})
		metricLogger.Debugf("got %d metrics for time range %s to %s, storing them into Presto into table %s", len(metrics), queryBegin, queryEnd, c.cfg.PrestoTableName)
		err := StorePrometheusMetrics(ctx, c.prestoQueryer, c.cfg.PrestoTableName, metrics)
		if err != nil {
			return fmt.Errorf("failed to store Prometheus metrics into table %s for the range %v to %v: %v",
				c.cfg.PrestoTableName, queryBegin, queryEnd, err)
		}
		metricLogger.Debugf("stored %d metrics for time range into Presto table %s successfully", len(metrics), c.cfg.PrestoTableName)
	} ***REMOVED*** {
		c.logger.Debugf("got 0 metrics for time range %s to %s", queryBegin, queryEnd)
	}
	c.metricsCount += len(metrics)
	return nil
}

func (c *PrometheusImporter) postProcessingHandler(_ context.Context, timeRanges []prom.Range) error {
	if len(timeRanges) != 0 {
		begin := timeRanges[0].Start.UTC()
		end := timeRanges[len(timeRanges)-1].End.UTC()
		c.logger.Infof("stored a total of %d metrics for data between %s and %s into %s", c.metricsCount, begin, end, c.cfg.PrestoTableName)
	} ***REMOVED*** {
		c.logger.Infof("no time ranges processed for %s", c.cfg.PrestoTableName)
	}
	return nil
}

func (c *PrometheusImporter) UpdateCon***REMOVED***g(cfg Con***REMOVED***g) {
	c.importLock.Lock()
	c.cfg = cfg
	c.logger = c.logger.WithFields(logrus.Fields{
		"tableName": cfg.PrestoTableName,
		"chunkSize": cfg.ChunkSize,
		"stepSize":  cfg.StepSize,
	})
	c.importLock.Unlock()
}

// ImportFromLastTimestamp executes a Presto query from the last time range it
// queried and stores the results in a Presto table.

// The importer will track the last time series it retrieved and will query
// the next time range starting from where it left off if paused or stopped.
// For more details on how querying Prometheus is done, see the package
// pkg/promquery.
func (c *PrometheusImporter) ImportFromLastTimestamp(ctx context.Context, allowIncompleteChunks bool) ([]prom.Range, error) {
	c.importLock.Lock()
	c.logger.Debugf("PrometheusImporter ImportFromLastTimestamp started")
	defer c.logger.Debugf("PrometheusImporter ImportFromLastTimestamp ***REMOVED***nished")
	defer c.importLock.Unlock()

	endTime := c.clock.Now().UTC()

	// if c.lastTimestamp is null then it's because we errored sometime
	// last time we collected and need to re-query Presto to ***REMOVED***gure out
	// the last timestamp
	if c.lastTimestamp == nil {
		var err error
		c.logger.Debugf("lastTimestamp for table %s: isn't known, querying for timestamp", c.cfg.PrestoTableName)
		c.lastTimestamp, err = getLastTimestampForTable(c.prestoQueryer, c.cfg.PrestoTableName)
		if err != nil {
			c.logger.WithError(err).Errorf("unable to get last timestamp for table %s", c.cfg.PrestoTableName)
			return nil, err
		}
	}

	var startTime time.Time
	if c.lastTimestamp != nil {
		c.logger.Debugf("lastTimestamp for table %s: %s", c.cfg.PrestoTableName, c.lastTimestamp.String())

		// We don't want to duplicate the c.lastTimestamp metric so add
		// the step size so that we start at the next interval no longer in
		// our range.
		startTime = c.lastTimestamp.Add(c.cfg.StepSize)
	} ***REMOVED*** {
		// Looks like we haven't populated any data in this table yet.
		// Let's back***REMOVED***ll our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == endTime, so it won't be queried, so this gets the chunk
		// before the latest
		startTime = endTime.Add(-2 * c.cfg.ChunkSize)
		c.logger.Debugf("no data in data store %s yet", c.cfg.PrestoTableName)
	}

	// If the startTime is too far back, we should limit this run to
	// maxChunkDuration so that if we're stopped for an extended amount of time,
	// this function won't return a slice with too many time ranges.
	totalChunkDuration := startTime.Sub(endTime)
	if totalChunkDuration >= maxChunkDuration {
		endTime = startTime.Add(maxChunkDuration)
	}

	return c.importMetrics(ctx, startTime, endTime, allowIncompleteChunks)
}

func (c *PrometheusImporter) ImportMetrics(ctx context.Context, startTime, endTime time.Time, allowIncompleteChunks bool) ([]prom.Range, error) {
	c.importLock.Lock()
	c.logger.Debugf("PrometheusImporter Import started")
	defer c.logger.Debugf("PrometheusImporter Import ***REMOVED***nished")
	defer c.importLock.Unlock()

	return c.importMetrics(ctx, startTime, endTime, allowIncompleteChunks)
}

func (c *PrometheusImporter) importMetrics(ctx context.Context, startTime, endTime time.Time, allowIncompleteChunks bool) ([]prom.Range, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"startTime": startTime,
		"endTime":   endTime,
	})
	queryRangeDuration := endTime.Sub(startTime)
	if c.cfg.MaxQueryRangeDuration != 0 && queryRangeDuration > c.cfg.MaxQueryRangeDuration {
		newEndTime := startTime.Add(c.cfg.MaxQueryRangeDuration)
		logger.Warnf("time range %s to %s exceeds PrometheusImporter MaxQueryRangeDuration %s, newEndTime: %s", startTime, endTime, c.cfg.MaxQueryRangeDuration, newEndTime)
		endTime = newEndTime
	}

	collectHandlers := promquery.ResultHandler{
		PreProcessingHandler:  c.preProcessingHandler,
		PreQueryHandler:       c.preQueryHandler,
		PostQueryHandler:      c.postQueryHandler,
		PostProcessingHandler: c.postProcessingHandler,
	}

	timeRanges, err := promquery.QueryRangeChunked(ctx, c.promConn, c.cfg.PrometheusQuery, startTime, endTime, c.cfg.ChunkSize, c.cfg.StepSize, c.cfg.MaxTimeRanges, allowIncompleteChunks, collectHandlers)
	if err != nil {
		logger.WithError(err).Error("error collecting metrics")
		// at this point we cannot be sure what is in Presto and what
		// isn't, so reset our c.lastTimestamp
		c.lastTimestamp = nil
		return timeRanges, err
	}

	if len(timeRanges) != 0 {
		lastTS := timeRanges[len(timeRanges)-1].End
		c.lastTimestamp = &lastTS
	}

	return timeRanges, nil
}

func promMatrixToPrometheusMetrics(timeRange prom.Range, matrix model.Matrix) []*PrometheusMetric {
	var metrics []*PrometheusMetric
	// iterate over segments of contiguous billing metrics
	for _, sampleStream := range matrix {
		for _, value := range sampleStream.Values {
			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			metric := &PrometheusMetric{
				Labels:    labels,
				Amount:    float64(value.Value),
				StepSize:  timeRange.Step,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

package prestostore

import (
	"context"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"

	"github.com/operator-framework/operator-metering/pkg/presto"
)

const (
	// cap the maximum importer.cfg.ChunkSize
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

func (importer *PrometheusImporter) UpdateCon***REMOVED***g(cfg Con***REMOVED***g) {
	importer.importLock.Lock()
	importer.cfg = cfg
	importer.logger = importer.logger.WithFields(logrus.Fields{
		"tableName": cfg.PrestoTableName,
		"chunkSize": cfg.ChunkSize,
		"stepSize":  cfg.StepSize,
	})
	importer.importLock.Unlock()
}

// ImportFromLastTimestamp executes a Presto query from the last time range it
// queried and stores the results in a Presto table.
// The importer will track the last time series it retrieved and will query
// the next time range starting from where it left off if paused or stopped.
// For more details on how querying Prometheus is done, see the package
// pkg/promquery.
func (importer *PrometheusImporter) ImportFromLastTimestamp(ctx context.Context, allowIncompleteChunks bool) ([]prom.Range, error) {
	importer.importLock.Lock()
	importer.logger.Debugf("PrometheusImporter ImportFromLastTimestamp started")
	defer importer.logger.Debugf("PrometheusImporter ImportFromLastTimestamp ***REMOVED***nished")
	defer importer.importLock.Unlock()

	endTime := importer.clock.Now().UTC()

	// if importer.lastTimestamp is null then it's because we errored sometime
	// last time we collected and need to re-query Presto to ***REMOVED***gure out
	// the last timestamp
	if importer.lastTimestamp == nil {
		var err error
		importer.logger.Debugf("lastTimestamp for table %s: isn't known, querying for timestamp", importer.cfg.PrestoTableName)
		importer.lastTimestamp, err = getLastTimestampForTable(importer.prestoQueryer, importer.cfg.PrestoTableName)
		if err != nil {
			importer.logger.WithError(err).Errorf("unable to get last timestamp for table %s", importer.cfg.PrestoTableName)
			return nil, err
		}
	}

	var startTime time.Time
	if importer.lastTimestamp != nil {
		importer.logger.Debugf("lastTimestamp for table %s: %s", importer.cfg.PrestoTableName, importer.lastTimestamp.String())

		// We don't want to duplicate the importer.lastTimestamp metric so add
		// the step size so that we start at the next interval no longer in
		// our range.
		startTime = importer.lastTimestamp.Add(importer.cfg.StepSize)
	} ***REMOVED*** {
		// Looks like we haven't populated any data in this table yet.
		// Let's back***REMOVED***ll our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == endTime, so it won't be queried, so this gets the chunk
		// before the latest
		startTime = endTime.Add(-2 * importer.cfg.ChunkSize)
		importer.logger.Debugf("no data in data store %s yet", importer.cfg.PrestoTableName)
	}

	// If the startTime is too far back, we should limit this run to
	// maxChunkDuration so that if we're stopped for an extended amount of time,
	// this function won't return a slice with too many time ranges.
	totalChunkDuration := startTime.Sub(endTime)
	if totalChunkDuration >= maxChunkDuration {
		endTime = startTime.Add(maxChunkDuration)
	}

	return importer.importMetrics(ctx, startTime, endTime, allowIncompleteChunks)
}

func (importer *PrometheusImporter) ImportMetrics(ctx context.Context, startTime, endTime time.Time, allowIncompleteChunks bool) ([]prom.Range, error) {
	importer.importLock.Lock()
	importer.logger.Debugf("PrometheusImporter Import started")
	defer importer.logger.Debugf("PrometheusImporter Import ***REMOVED***nished")
	defer importer.importLock.Unlock()

	return importer.importMetrics(ctx, startTime, endTime, allowIncompleteChunks)
}

func (importer *PrometheusImporter) importMetrics(ctx context.Context, startTime, endTime time.Time, allowIncompleteChunks bool) ([]prom.Range, error) {
	logger := importer.logger.WithFields(logrus.Fields{
		"startTime": startTime,
		"endTime":   endTime,
	})
	queryRangeDuration := endTime.Sub(startTime)
	if importer.cfg.MaxQueryRangeDuration != 0 && queryRangeDuration > importer.cfg.MaxQueryRangeDuration {
		newEndTime := startTime.Add(importer.cfg.MaxQueryRangeDuration)
		logger.Warnf("time range %s to %s exceeds PrometheusImporter MaxQueryRangeDuration %s, newEndTime: %s", startTime, endTime, importer.cfg.MaxQueryRangeDuration, newEndTime)
		endTime = newEndTime
	}

	timeRanges, err := importer.importFromTimeRange(ctx, startTime, endTime, allowIncompleteChunks)

	if err != nil {
		logger.WithError(err).Error("error collecting metrics")
		// at this point we cannot be sure what is in Presto and what
		// isn't, so reset our importer.lastTimestamp
		importer.lastTimestamp = nil
		return timeRanges, err
	}

	if len(timeRanges) != 0 {
		lastTS := timeRanges[len(timeRanges)-1].End
		importer.lastTimestamp = &lastTS
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

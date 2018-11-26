package prestostore

import (
	"context"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
)

type ImporterMetricsCollectors struct {
	TotalImportsCounter     prometheus.Counter
	FailedImportsCounter    prometheus.Counter
	ImportDurationHistogram prometheus.Observer

	TotalPrometheusQueriesCounter    prometheus.Counter
	FailedPrometheusQueriesCounter   prometheus.Counter
	PrometheusQueryDurationHistogram prometheus.Observer

	TotalPrestoStoresCounter     prometheus.Counter
	FailedPrestoStoresCounter    prometheus.Counter
	PrestoStoreDurationHistogram prometheus.Observer

	MetricsScrapedCounter  prometheus.Counter
	MetricsImportedCounter prometheus.Counter

	ImportsRunningGauge prometheus.Gauge
}

// PrometheusImporter imports Prometheus metrics into Presto tables
type PrometheusImporter struct {
	logger                logrus.FieldLogger
	promConn              prom.API
	prometheusMetricsRepo PrometheusMetricsRepo
	clock                 clock.Clock
	cfg                   Con***REMOVED***g

	metricsCollectors ImporterMetricsCollectors

	// importLock ensures only one import is running at a time, protecting the
	// lastTimestamp and metrics ***REMOVED***elds
	importLock sync.Mutex

	// lastTimestamp is the lastTimestamp stored for this PrometheusImporter
	lastTimestamp *time.Time
}

type Con***REMOVED***g struct {
	PrometheusQuery           string
	PrestoTableName           string
	ChunkSize                 time.Duration
	StepSize                  time.Duration
	MaxTimeRanges             int64
	MaxQueryRangeDuration     time.Duration
	ImportFromTime            *time.Time
	MaxBack***REMOVED***llImportDuration time.Duration
}

func NewPrometheusImporter(logger logrus.FieldLogger, promConn prom.API, prometheusMetricsRepo PrometheusMetricsRepo, clock clock.Clock, cfg Con***REMOVED***g, collectors ImporterMetricsCollectors) *PrometheusImporter {
	logger = logger.WithFields(logrus.Fields{
		"component": "PrometheusImporter",
		"tableName": cfg.PrestoTableName,
		"chunkSize": cfg.ChunkSize,
		"stepSize":  cfg.StepSize,
	})
	return &PrometheusImporter{
		logger:                logger,
		promConn:              promConn,
		prometheusMetricsRepo: prometheusMetricsRepo,
		clock:             clock,
		cfg:               cfg,
		metricsCollectors: collectors,
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
func (importer *PrometheusImporter) ImportFromLastTimestamp(ctx context.Context, allowIncompleteChunks bool) (*PrometheusImportResults, error) {
	importer.importLock.Lock()
	importer.logger.Debugf("PrometheusImporter ImportFromLastTimestamp started")
	defer importer.logger.Debugf("PrometheusImporter ImportFromLastTimestamp ***REMOVED***nished")
	defer importer.importLock.Unlock()

	endTime := importer.clock.Now().UTC()

	cfg := importer.cfg

	// if importer.lastTimestamp is null then it's because we haven't run
	// before, we have been restarted (error, or not) and do not know the
	// last time we collected and need to re-query Presto to ***REMOVED***gure out
	// the last timestamp
	if importer.lastTimestamp == nil {
		var err error
		importer.logger.Debugf("lastTimestamp for table %s: isn't known, querying for timestamp", cfg.PrestoTableName)
		importer.lastTimestamp, err = importer.prometheusMetricsRepo.GetLastTimestampForTable(cfg.PrestoTableName)
		if err != nil {
			importer.logger.WithError(err).Errorf("unable to get last timestamp for table %s", cfg.PrestoTableName)
			return nil, err
		}
	}

	var startTime time.Time
	// if lastTimestamp is still nil, but we didn't error than there is no
	// last timestamp and this is the ***REMOVED***rst collection, if not then our query
	// above found a timestamp in the table
	if importer.lastTimestamp != nil {
		importer.logger.Debugf("lastTimestamp for table %s: %s", cfg.PrestoTableName, importer.lastTimestamp.String())
		// We don't want to duplicate the importer.lastTimestamp metric so add
		// the step size so that we start at the next interval no longer in
		// our range.
		startTime = importer.lastTimestamp.Add(cfg.StepSize)
	} ***REMOVED*** {
		// check if we're supposed to start from a speci***REMOVED***c
		// time, and if not back***REMOVED***ll a default amount
		if cfg.ImportFromTime != nil {
			importer.logger.Debugf("importFromTimestamp for table %s: %s", cfg.PrestoTableName, cfg.ImportFromTime.String())
			startTime = *cfg.ImportFromTime
		} ***REMOVED*** {
			importer.logger.Debugf("no lastTimestamp or importFromTime for table %s: back***REMOVED***lling %s", cfg.PrestoTableName, cfg.MaxBack***REMOVED***llImportDuration)
			startTime = endTime.Add(-cfg.MaxBack***REMOVED***llImportDuration)
		}
		importer.logger.Infof("no data in table %s: back***REMOVED***lling from %s until %s", cfg.PrestoTableName, startTime, endTime)
	}

	// If the startTime is too far back, we should limit this run to
	// cfg.MaxQueryRangeDuration so that if we're stopped for an
	// extended amount of time, this function won't return a slice with too
	// many time ranges.
	totalChunkDuration := startTime.Sub(endTime)
	if totalChunkDuration >= cfg.MaxQueryRangeDuration {
		endTime = startTime.Add(cfg.MaxQueryRangeDuration)
	}

	importResults, err := ImportFromTimeRange(importer.logger, importer.clock, importer.promConn, importer.prometheusMetricsRepo, importer.metricsCollectors, ctx, startTime, endTime, cfg, allowIncompleteChunks)
	if err != nil {
		importer.logger.WithError(err).Error("error collecting metrics")
		// at this point we cannot be sure what is in Presto and what
		// isn't, so reset our importer.lastTimestamp
		importer.lastTimestamp = nil
		return &importResults, err
	}

	if len(importResults.ProcessedTimeRanges) != 0 {
		lastTS := importResults.ProcessedTimeRanges[len(importResults.ProcessedTimeRanges)-1].End
		importer.lastTimestamp = &lastTS
	}
	return &importResults, nil
}

func promMatrixToPrometheusMetrics(timeRange prom.Range, matrix model.Matrix) []*PrometheusMetric {
	var metrics []*PrometheusMetric
	// iterate over segments of contiguous billing metrics
	for _, sampleStream := range matrix {
		labels := make(map[string]string, len(sampleStream.Metric))
		for k, v := range sampleStream.Metric {
			labels[string(k)] = string(v)
		}
		for _, value := range sampleStream.Values {
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

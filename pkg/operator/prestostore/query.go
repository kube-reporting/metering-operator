package prestostore

import (
	"context"
	"fmt"
	"sort"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
)

type PrometheusImportResults struct {
	ProcessedTimeRanges []prom.Range
	Metrics             []*PrometheusMetric
}

// importFromTimeRange executes a promQL query over the interval between start
// and end, performing multiple Prometheus query_range queries of chunkSize.
// Returns the time ranges queried and any errors encountered. Stops after the
// first error, consult timeRanges to determine how many chunks were queried.
//
// If the number of queries exceeds maxTimeRanges, then the timeRanges
// exceeding that count will be skipped. The allowIncompleteChunks parameter
// controls whether or not every chunk must be a full chunkSize, or if there
// can be incomplete chunks. This has an effect when there is only one chunk
// that's incomplete, and if there are multiple chunks, whether or not the
// final chunk up to the endTime will be included even if the duration of
// endTime - startTime isn't perfectly divisible by chunkSize.
func ImportFromTimeRange(logger logrus.FieldLogger, clock clock.Clock, promConn prom.API, prometheusMetricsStorer PrometheusMetricsStorer, metricsCollectors ImporterMetricsCollectors, ctx context.Context, startTime, endTime time.Time, cfg Config, allowIncompleteChunks bool) (PrometheusImportResults, error) {
	metricsCollectors.ImportsRunningGauge.Inc()

	queryRangeDuration := endTime.Sub(startTime)
	if cfg.MaxQueryRangeDuration != 0 && queryRangeDuration > cfg.MaxQueryRangeDuration {
		newEndTime := startTime.Add(cfg.MaxQueryRangeDuration)
		logger.Warnf("time range %s to %s exceeds PrometheusImporter MaxQueryRangeDuration %s, newEndTime: %s", startTime, endTime, cfg.MaxQueryRangeDuration, newEndTime)
		endTime = newEndTime
	}

	importStart := clock.Now()
	metricsCollectors.TotalImportsCounter.Inc()

	defer func() {
		metricsCollectors.ImportsRunningGauge.Dec()
		importDuration := clock.Since(importStart)
		metricsCollectors.ImportDurationHistogram.Observe(importDuration.Seconds())
		logger.Debugf("took %s to run import", importDuration)
	}()

	timeRanges := getTimeRangesChunked(startTime, endTime, cfg.ChunkSize, cfg.StepSize, cfg.MaxTimeRanges, allowIncompleteChunks)

	var importResults PrometheusImportResults
	if len(timeRanges) == 0 {
		logger.Debugf("no time ranges to query yet for table %s", cfg.PrestoTableName)
		return importResults, nil
	}

	metricsCount := 0
	startTime = timeRanges[0].Start.UTC()
	endTime = timeRanges[len(timeRanges)-1].End.UTC()
	logger = logger.WithFields(logrus.Fields{
		"startTime": startTime,
		"endTime":   endTime,
	})
	logger.Debugf("querying for data between %s and %s (chunks: %d)", startTime, endTime, len(timeRanges))

	for _, timeRange := range timeRanges {
		// check for cancellation
		select {
		case <-ctx.Done():
			return importResults, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		promQueryBegin := timeRange.Start.UTC()
		promQueryEnd := timeRange.End.UTC()
		promLogger := logger.WithFields(logrus.Fields{
			"promQueryBegin": promQueryBegin,
			"promQueryEnd":   promQueryEnd,
		})

		promLogger.Debugf("querying Prometheus using range %s to %s", timeRange.Start, timeRange.End)

		queryStart := clock.Now()
		pVal, err := promConn.QueryRange(ctx, cfg.PrometheusQuery, timeRange)
		queryDuration := clock.Since(queryStart)
		metricsCollectors.PrometheusQueryDurationHistogram.Observe(float64(queryDuration.Seconds()))
		metricsCollectors.TotalPrometheusQueriesCounter.Inc()
		if err != nil {
			metricsCollectors.FailedImportsCounter.Inc()
			metricsCollectors.FailedPrometheusQueriesCounter.Inc()
			return importResults, fmt.Errorf("failed to perform Prometheus query: %v", err)
		}

		matrix, ok := pVal.(model.Matrix)
		if !ok {
			return importResults, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
		}

		metrics := promMatrixToPrometheusMetrics(timeRange, matrix)
		numMetrics := len(metrics)
		metricsCollectors.MetricsScrapedCounter.Add(float64(numMetrics))

		// check for cancellation
		select {
		case <-ctx.Done():
			return importResults, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		if numMetrics != 0 {
			metricsBegin := metrics[0].Timestamp
			metricsEnd := metrics[numMetrics-1].Timestamp
			logger := promLogger.WithFields(logrus.Fields{
				"metricsBegin": metricsBegin,
				"metricsEnd":   metricsEnd,
			})
			logger.Debugf("got %d metrics for time range %s to %s, storing them into Presto into table %s", numMetrics, promQueryBegin, promQueryEnd, cfg.PrestoTableName)

			metricsCollectors.TotalPrometheusQueriesCounter.Inc()
			prestoStoreBegin := clock.Now()
			err := prometheusMetricsStorer.StorePrometheusMetrics(ctx, cfg.PrestoTableName, metrics)
			prestoStoreDuration := clock.Since(prestoStoreBegin)
			metricsCollectors.PrestoStoreDurationHistogram.Observe(float64(prestoStoreDuration.Seconds()))
			if err != nil {
				metricsCollectors.FailedImportsCounter.Inc()
				metricsCollectors.FailedPrestoStoresCounter.Inc()
				return importResults, fmt.Errorf("failed to store Prometheus metrics into table %s for the range %v to %v: %v",
					cfg.PrestoTableName, promQueryBegin, promQueryEnd, err)
			}
			// Ensure the metrics are sorted by timestamp
			sort.Slice(metrics, func(i, j int) bool {
				return metrics[i].Timestamp.Before(metrics[j].Timestamp)
			})
			importResults.Metrics = metrics
			logger.Debugf("stored %d metrics for time range %s to %s into Presto table %s (took %s)", numMetrics, promQueryBegin, promQueryEnd, cfg.PrestoTableName, prestoStoreDuration)
			metricsCollectors.MetricsImportedCounter.Add(float64(numMetrics))
			metricsCount += numMetrics
		}

		importResults.ProcessedTimeRanges = append(importResults.ProcessedTimeRanges, timeRange)
	}

	if len(importResults.ProcessedTimeRanges) != 0 {
		begin := importResults.ProcessedTimeRanges[0].Start.UTC()
		end := importResults.ProcessedTimeRanges[len(timeRanges)-1].End.UTC()
		logger.Infof("stored a total of %d metrics for data between %s and %s into %s", metricsCount, begin, end, cfg.PrestoTableName)
		return importResults, nil
	} else {
		logger.Infof("no time ranges processed for %s", cfg.PrestoTableName)
		return importResults, nil
	}
}

func getTimeRangesChunked(beginTime, endTime time.Time, chunkSize, stepSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool) []prom.Range {
	chunkStart := truncateToSecond(beginTime)
	chunkEnd := truncateToSecond(chunkStart.Add(chunkSize))

	// don't set a limit if negative or zero
	disableMax := maxTimeRanges <= 0

	var timeRanges []prom.Range
	for i := int64(0); disableMax || (i < maxTimeRanges); i++ {
		if allowIncompleteChunks {
			if chunkEnd.After(endTime) {
				chunkEnd = truncateToSecond(endTime)
			}
			if chunkEnd.Equal(chunkStart) {
				break
			}
		} else {
			// Do not collect data after endTime
			if chunkEnd.After(endTime) {
				break
			}

			// Only get chunks that are a full chunk size
			if chunkEnd.Sub(chunkStart) < chunkSize {
				break
			}
		}
		timeRanges = append(timeRanges, prom.Range{
			Start: chunkStart.UTC(),
			End:   chunkEnd.UTC(),
			Step:  stepSize,
		})

		if allowIncompleteChunks && chunkEnd.Equal(truncateToSecond(endTime)) {
			break
		}

		// Add the metrics step size to the start time so that we don't
		// re-query the Previous ranges end time in this range
		chunkStart = truncateToSecond(chunkEnd.Add(stepSize))
		// Add chunkSize to the end time to get our full chunk. If the chunkEnd
		// is past the endTime, then this chunk is skipped.
		chunkEnd = truncateToSecond(chunkStart.Add(chunkSize))
	}

	return timeRanges
}

func truncateToSecond(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

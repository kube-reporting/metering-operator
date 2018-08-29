package prestostore

import (
	"context"
	"fmt"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

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
func (importer *PrometheusImporter) importFromTimeRange(ctx context.Context, startTime, endTime time.Time, allowIncompleteChunks bool) ([]prom.Range, error) {
	timeRanges := getTimeRangesChunked(startTime, endTime, importer.cfg.ChunkSize, importer.cfg.StepSize, importer.cfg.MaxTimeRanges, allowIncompleteChunks)
	var processedTimeRanges []prom.Range
	metricsCount := 0

	if len(timeRanges) == 0 {
		importer.logger.Infof("no time ranges to query yet for table %s", importer.cfg.PrestoTableName)
		return nil, nil
	} else {
		begin := timeRanges[0].Start.UTC()
		end := timeRanges[len(timeRanges)-1].End.UTC()
		logger := importer.logger.WithFields(logrus.Fields{
			"rangeBegin": begin,
			"rangeEnd":   end,
		})
		logger.Debugf("querying for data between %s and %s (chunks: %d)", begin, end, len(timeRanges))
	}

	for _, timeRange := range timeRanges {
		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		promQueryBegin := timeRange.Start.UTC()
		promQueryEnd := timeRange.End.UTC()
		promLogger := importer.logger.WithFields(logrus.Fields{
			"promQueryBegin": promQueryBegin,
			"promQueryEnd":   promQueryEnd,
		})

		promLogger.Debugf("querying Prometheus using range %s to %s", timeRange.Start, timeRange.End)

		pVal, err := importer.promConn.QueryRange(ctx, importer.cfg.PrometheusQuery, timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to perform Prometheus query: %v", err)
		}

		matrix, ok := pVal.(model.Matrix)
		if !ok {
			return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
		}

		metrics := promMatrixToPrometheusMetrics(timeRange, matrix)
		numMetrics := len(metrics)

		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
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
			logger.Debugf("got %d metrics for time range %s to %s, storing them into Presto into table %s", numMetrics, promQueryBegin, promQueryEnd, importer.cfg.PrestoTableName)

			err := StorePrometheusMetrics(ctx, importer.prestoQueryer, importer.cfg.PrestoTableName, metrics)
			if err != nil {
				return nil, fmt.Errorf("failed to store Prometheus metrics into table %s for the range %v to %v: %v",
					importer.cfg.PrestoTableName, promQueryBegin, promQueryEnd, err)
			}
			logger.Debugf("stored %d metrics for time range %s to %s into Presto table %s", numMetrics, promQueryBegin, promQueryEnd, importer.cfg.PrestoTableName)
			metricsCount += numMetrics
		}

		processedTimeRanges = append(processedTimeRanges, timeRange)
	}

	if len(processedTimeRanges) != 0 {
		begin := processedTimeRanges[0].Start.UTC()
		end := processedTimeRanges[len(timeRanges)-1].End.UTC()
		importer.logger.Infof("stored a total of %d metrics for data between %s and %s into %s", metricsCount, begin, end, importer.cfg.PrestoTableName)
	} else {
		importer.logger.Infof("no time ranges processed for %s", importer.cfg.PrestoTableName)
	}
	return processedTimeRanges, nil
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

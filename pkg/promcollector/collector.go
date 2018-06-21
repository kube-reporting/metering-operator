package promcollector

import (
	"context"
	"fmt"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Collector queries Prometheus and handles querying Prometheus time series
// over a given time range, breaking the queries up into multiple "chunks"
// of a specific duration, to reduce the number of results being returned from
// Prometheus at any given time.
type Collector struct {
	promConn              prom.API
	query                 string
	preProcessingHandler  func(context.Context, []prom.Range) error
	preQueryHandler       func(context.Context, prom.Range) error
	postQueryHandler      func(context.Context, prom.Range, model.Matrix) error
	postProcessingHandler func(context.Context, []prom.Range) error
}

func New(
	promConn prom.API,
	query string,
	preProcessingHandler func(context.Context, []prom.Range) error,
	preQueryHandler func(context.Context, prom.Range) error,
	postQueryHandler func(context.Context, prom.Range, model.Matrix) error,
	postProcessingHandler func(context.Context, []prom.Range) error,
) *Collector {
	return &Collector{
		query:                 query,
		promConn:              promConn,
		preProcessingHandler:  preProcessingHandler,
		preQueryHandler:       preQueryHandler,
		postQueryHandler:      postQueryHandler,
		postProcessingHandler: postProcessingHandler,
	}
}

// Collect runs the configured query over the interval between start and end,
// performing multiple Prometheus query_range queries of chunkSize. Returns the
// time ranges queried and any errors encountered. Stops after the first error,
// consult timeRanges to determine how many chunks were queried.
//
// If the number of queries exceeds maxTimeRanges, then the timeRanges
// exceeding that count will be skipped. The allowIncompleteChunks parameter
// controls whether or not every chunk must be a full chunkSize, or if there
// can be incomplete chunks. This has an effect when there is only one chunk
// that's incomplete, and if there are multiple chunks, whether or not the
// final chunk up to the endTime will be included even if the duration of
// endTime - startTime isn't perfectly divisible by chunkSize.
func (c *Collector) Collect(ctx context.Context, startTime, endTime time.Time, stepSize, chunkSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool) (timeRanges []prom.Range, err error) {
	timeRangesToProcess := getTimeRanges(startTime, endTime, chunkSize, stepSize, maxTimeRanges, allowIncompleteChunks)
	if len(timeRangesToProcess) == 0 {
		return nil, nil
	}

	if c.preProcessingHandler != nil {
		err = c.preProcessingHandler(ctx, timeRangesToProcess)
		if err != nil {
			return timeRanges, err
		}
	}

	for _, timeRange := range timeRangesToProcess {
		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		if c.preQueryHandler != nil {
			err = c.preQueryHandler(ctx, timeRange)
			if err != nil {
				return timeRanges, err
			}
		}

		pVal, err := c.promConn.QueryRange(ctx, c.query, timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to perform Prometheus query: %v", err)
		}

		matrix, ok := pVal.(model.Matrix)
		if !ok {
			return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
		}

		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		if c.postQueryHandler != nil {
			err = c.postQueryHandler(ctx, timeRange, matrix)
			if err != nil {
				return timeRanges, err
			}
		}
		timeRanges = append(timeRanges, timeRange)
	}

	if c.postProcessingHandler != nil {
		err = c.postProcessingHandler(ctx, timeRanges)
		if err != nil {
			return timeRanges, err
		}
	}
	return timeRanges, nil
}

func getTimeRanges(beginTime, endTime time.Time, chunkSize, stepSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool) []prom.Range {
	chunkStart := truncateToMinute(beginTime)
	chunkEnd := truncateToMinute(chunkStart.Add(chunkSize))

	// don't set a limit if negative or zero
	disableMax := maxTimeRanges <= 0

	var timeRanges []prom.Range
	for i := int64(0); disableMax || (i < maxTimeRanges); i++ {
		if allowIncompleteChunks {
			if chunkEnd.After(endTime) {
				chunkEnd = truncateToMinute(endTime)
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

		if allowIncompleteChunks && chunkEnd.Equal(truncateToMinute(endTime)) {
			break
		}

		// Add the metrics step size to the start time so that we don't
		// re-query the previous ranges end time in this range
		chunkStart = truncateToMinute(chunkEnd.Add(stepSize))
		// Add chunkSize to the end time to get our full chunk. If the chunkEnd
		// is past the endTime, then this chunk is skipped.
		chunkEnd = truncateToMinute(chunkStart.Add(chunkSize))
	}

	return timeRanges
}

func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

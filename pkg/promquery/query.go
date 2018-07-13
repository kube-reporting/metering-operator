package promquery

import (
	"context"
	"fmt"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type ResultHandler struct {
	PreProcessingHandler  func(context.Context, []prom.Range) error
	PreQueryHandler       func(context.Context, prom.Range) error
	PostQueryHandler      func(context.Context, prom.Range, model.Matrix) error
	PostProcessingHandler func(context.Context, []prom.Range) error
}

// QueryRangeChunked executes a promQL query over the interval between start
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
func QueryRangeChunked(ctx context.Context, promConn prom.API, query string, startTime, endTime time.Time, chunkSize, stepSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool, handlers ResultHandler) (timeRanges []prom.Range, err error) {
	timeRangesToProcess := getTimeRanges(startTime, endTime, chunkSize, stepSize, maxTimeRanges, allowIncompleteChunks)

	if handlers.PreProcessingHandler != nil {
		err = handlers.PreProcessingHandler(ctx, timeRangesToProcess)
		if err != nil {
			return timeRanges, err
		}
	}

	if len(timeRangesToProcess) == 0 {
		return nil, nil
	}

	for _, timeRange := range timeRangesToProcess {
		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		if handlers.PreQueryHandler != nil {
			err = handlers.PreQueryHandler(ctx, timeRange)
			if err != nil {
				return timeRanges, err
			}
		}

		pVal, err := promConn.QueryRange(ctx, query, timeRange)
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

		if handlers.PostQueryHandler != nil {
			err = handlers.PostQueryHandler(ctx, timeRange, matrix)
			if err != nil {
				return timeRanges, err
			}
		}
		timeRanges = append(timeRanges, timeRange)
	}

	if handlers.PostProcessingHandler != nil {
		err = handlers.PostProcessingHandler(ctx, timeRanges)
		if err != nil {
			return timeRanges, err
		}
	}
	return timeRanges, nil
}

func getTimeRanges(beginTime, endTime time.Time, chunkSize, stepSize time.Duration, maxTimeRanges int64, allowIncompleteChunks bool) []prom.Range {
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

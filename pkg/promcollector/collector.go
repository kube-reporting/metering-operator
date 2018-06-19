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
// of a speci***REMOVED***c duration, to reduce the number of results being returned from
// Prometheus at any given time.
type Collector struct {
	promConn             prom.API
	query                string
	preProcessingHandler func(context.Context, []prom.Range) error
	preQueryHandler      func(context.Context, prom.Range) error
	postQueryHandler     func(context.Context, prom.Range, []*Record) error
}

func New(
	promConn prom.API,
	query string,
	preProcessingHandler func(context.Context, []prom.Range) error,
	preQueryHandler func(context.Context, prom.Range) error,
	postQueryHandler func(context.Context, prom.Range, []*Record) error,
) *Collector {
	return &Collector{
		query:                query,
		promConn:             promConn,
		preProcessingHandler: preProcessingHandler,
		preQueryHandler:      preQueryHandler,
		postQueryHandler:     postQueryHandler,
	}
}

// Collect runs the con***REMOVED***gured query over the interval between start and end,
// performing multiple Prometheus query_range queries of chunkSize. Returns the
// time ranges queried and any errors encountered. Stops after the ***REMOVED***rst error,
// consult timeRanges to determine how many chunks were queried.
//
// If the number of queries exceeds maxTimeRanges, then the timeRanges
// exceeding that count will be skipped. The allowIncompleteChunks parameter
// controls whether or not every chunk must be a full chunkSize, or if there
// can be incomplete chunks. This has an effect when there is only one chunk
// that's incomplete, and if there are multiple chunks, whether or not the
// ***REMOVED***nal chunk up to the endTime will be included even if the duration of
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

		records, err := Query(ctx, c.promConn, c.query, timeRange)
		if err != nil {
			return timeRanges, err
		}

		// check for cancellation
		select {
		case <-ctx.Done():
			return timeRanges, ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		if c.postQueryHandler != nil {
			err = c.postQueryHandler(ctx, timeRange, records)
			if err != nil {
				return timeRanges, err
			}
		}
		timeRanges = append(timeRanges, timeRange)
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
		} ***REMOVED*** {
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

// Record is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type Record struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}

func Query(ctx context.Context, promConn prom.API, query string, queryRng prom.Range) ([]*Record, error) {
	pVal, err := promConn.QueryRange(ctx, query, queryRng)
	if err != nil {
		return nil, fmt.Errorf("failed to perform billing query: %v", err)
	}

	matrix, ok := pVal.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
	}

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
				StepSize:  queryRng.Step,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}

func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

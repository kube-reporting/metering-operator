package prestostore

import (
	"testing"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/assert"
)

func TestGetTimeRanges(t *testing.T) {
	janOne := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		startTime      time.Time
		endTime        time.Time
		chunkSize      time.Duration
		stepSize       time.Duration
		maxTimeRanges  int64
		expectedRanges []prom.Range
	}{
		"start and end are zero": {
			chunkSize:      time.Minute * 5,
			stepSize:       time.Minute,
			expectedRanges: nil,
		},
		"start and end are same": {
			startTime:      janOne,
			endTime:        janOne,
			chunkSize:      time.Minute * 5,
			stepSize:       time.Minute,
			expectedRanges: nil,
		},
		"period is exactly divisible by chunkSize": {
			startTime: janOne,
			endTime:   janOne.Add(2 * time.Hour),
			chunkSize: time.Hour,
			stepSize:  time.Minute,
			expectedRanges: []prom.Range{
				{
					Start: janOne,
					End:   janOne.Add(time.Hour),
					Step:  time.Minute,
				},
				// There is no second chunk, because it would be too small with
				// stepSize added
			},
		},
		"period is divisible by chunkSize with stepSize added": {
			startTime: janOne,
			endTime:   janOne.Add(2 * time.Hour).Add(time.Minute), // Add stepSize
			chunkSize: time.Hour,
			stepSize:  time.Minute,
			expectedRanges: []prom.Range{
				{
					Start: janOne,
					End:   janOne.Add(time.Hour),
					Step:  time.Minute,
				},
				{
					Start: janOne.Add(time.Hour + time.Minute),
					End:   janOne.Add(2*time.Hour + time.Minute),
					Step:  time.Minute,
				},
			},
		},
		"period is less than divisible by chunkSize": {
			startTime:      janOne,
			endTime:        janOne.Add(30 * time.Minute),
			chunkSize:      time.Hour,
			stepSize:       time.Minute,
			expectedRanges: nil,
		},
	}

	for name, test := range tests {
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			timeRanges := getTimeRangesChunked(test.startTime, test.endTime, test.chunkSize, test.stepSize, test.maxTimeRanges)
			assert.Equal(t, test.expectedRanges, timeRanges)
		})
	}

}

package promsum

import (
	"time"
)

// Range is an uninterrupted period of time.
type Range struct {
	Start time.Time
	End   time.Time
}

// Within returns true if a the date given overlaps with this range.
func (r Range) Within(t time.Time) bool {
	if t.Before(r.Start) {
		return false
	} else if t.After(r.End) {
		return false
	}
	return true
}

// Segment divides the given ranges when a range exceeds the max period.
func Segment(rngs []Range, max time.Duration) ([]Range, error) {
	return nil, nil
}

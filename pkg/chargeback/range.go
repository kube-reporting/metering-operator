package chargeback

import (
	"fmt"
	"strconv"
	"time"
)

// Range is an uninterrupted period of time.
type Range struct {
	Start time.Time
	End   time.Time
}

// ParseUnixRange takes 2 strings containing representations of integer Unix timestamps. The timestamps must be valid.
func ParseUnixRange(startStr, endStr string) (Range, error) {
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return Range{}, fmt.Errorf("couldn't parse start of range '%s': %v", startStr, err)
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return Range{}, fmt.Errorf("couldn't parse end of range '%s': %v", endStr, err)
	}

	return Range{
		Start: time.Unix(start, 0),
		End:   time.Unix(end, 0),
	}, nil
}

// Within returns true if a the date given overlaps with this range.
func (r Range) Within(t time.Time) bool {
	if t.Before(r.Start) {
		return false
	} ***REMOVED*** if t.After(r.End) {
		return false
	}
	return true
}

// Equal returns true if both ranges are identical.
func (r Range) Equal(o Range) bool {
	equal := (r.Start.Equal(o.Start) && r.End.Equal(o.End))
	return equal
}

// String returns a human readable representation of a range.
func (r Range) String() string {
	return fmt.Sprintf("Range[%s to %s]", r.Start.Format(time.RFC3339), r.End.Format(time.RFC3339))
}

// Segment returns consecutive subranges the length of interval. If interval is larger then the remainder, the remainder
// is returned. If interval is <= 0, an empty range is returned.
func (r Range) Segment(interval time.Duration) (ranges []Range) {
	if interval <= 0 {
		return
	}

	start := r.Start
	for start.Before(r.End) {
		rng := Range{
			Start: start,
			End:   start.Add(interval),
		}

		// if rest of range is smaller than duration of segment, return full remainder
		if !rng.End.Before(r.End) {
			rng.End = r.End
		}

		ranges = append(ranges, rng)
		start = rng.End
	}
	return
}

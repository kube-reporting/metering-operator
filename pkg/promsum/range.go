package promsum

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
	} else if t.After(r.End) {
		return false
	}
	return true
}

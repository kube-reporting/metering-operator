package promsum

import (
	"strings"
	"testing"
	"time"
)

func TestParseUnixRange(t *testing.T) {
	expected := []struct {
		start string
		end   string
		rng   Range
		// errContains should be a substring of the error returned
		errContains string
	}{
		{
			start: "1497472873",
			end:   "1497478553",
			rng: Range{
				Start: time.Unix(1497472873, 0),
				End:   time.Unix(1497478553, 0),
			},
		},
		{
			start: "45",
			end:   "1000",
			rng: Range{
				Start: time.Unix(45, 0),
				End:   time.Unix(1000, 0),
			},
		},
		{
			start:       "14974 72873", // invalid
			end:         "1497478553",
			errContains: "start",
		},
		{
			start:       "45",
			end:         "1000b", // invalid
			errContains: "end",
		},
	}

	for _, e := range expected {
		rng, err := ParseUnixRange(e.start, e.end)
		// check errors
		if err != nil {
			if len(e.errContains) == 0 {
				t.Error("encountered unexpected error parsing unix time range: ", err)
			} ***REMOVED*** {
				expectStr, actualStr := strings.ToLower(e.errContains), strings.ToLower(err.Error())
				if !strings.Contains(actualStr, expectStr) {
					t.Errorf("error should contain '%s' instead got: %s", expectStr, err.Error())
				}
			}
			continue
		}

		// check return value is expected
		if !rng.Equal(e.rng) {
			t.Errorf("returned range different than expected - wanted: %v, got: %v", e.rng, rng)
		}
	}
}

func TestRange_Equal(t *testing.T) {
	expected := []struct {
		***REMOVED***rst  Range
		second Range
		equal  bool
	}{
		{
			***REMOVED***rst: Range{
				Start: time.Unix(8, 0),
				End:   time.Unix(30, 0),
			},
			second: Range{
				Start: time.Unix(8, 0),
				End:   time.Unix(30, 0),
			},
			equal: true,
		},
		{
			***REMOVED***rst: Range{
				Start: time.Unix(8, 0),
				End:   time.Unix(30, 0),
			},
			second: Range{
				Start: time.Unix(8, 0),
				End:   time.Unix(60, 0),
			},
			equal: false,
		},
		{
			***REMOVED***rst: Range{
				Start: time.Unix(-30, 0),
				End:   time.Unix(30, 0),
			},
			second: Range{
				Start: time.Unix(8, 0),
				End:   time.Unix(30, 0),
			},
			equal: false,
		},
	}

	for _, e := range expected {
		actualEqual := e.***REMOVED***rst.Equal(e.second)
		if actualEqual != e.equal {
			t.Errorf("unexpected result: wanted %v, got %v", e.equal, actualEqual)
		}
	}
}

func TestRange_Within(t *testing.T) {
	rng := Range{
		Start: time.Unix(100, 0),
		End:   time.Unix(200, 0),
	}
	expected := []struct {
		time.Time
		Within bool
	}{
		{ // before
			Time: time.Unix(50, 0),
		},
		{ // after
			Time: time.Unix(250, 0),
		},
		{ // start
			Time:   time.Unix(100, 0),
			Within: true,
		},
		{ // end
			Time:   time.Unix(200, 0),
			Within: true,
		},
		{ // inside
			Time:   time.Unix(150, 0),
			Within: true,
		},
	}

	for _, e := range expected {
		actualWithin := rng.Within(e.Time)
		if actualWithin != e.Within {
			t.Errorf("expected different result for whether %s is within %v: wanted %v, got %v",
				e.Time.Format(time.RFC3339), rng, e.Within, actualWithin)
		}
	}
}

func TestRange_String(t *testing.T) {
	// test con***REMOVED***rm String function uses RFC3339 times
	now := time.Now().UTC()
	rng := Range{
		Start: now,
		End:   now.Add(48 * time.Hour),
	}

	rngStr := rng.String()
	if !strings.Contains(rngStr, rng.Start.Format(time.RFC3339)) {
		t.Errorf("did not ***REMOVED***nd RFC3339 (%s) time for start, got: %s", rng.Start.Format(time.RFC3339), rngStr)
	} ***REMOVED*** if !strings.Contains(rngStr, rng.End.Format(time.RFC3339)) {
		t.Errorf("did not ***REMOVED***nd RFC3339 (%s) time for end, got: %s", rng.End.Format(time.RFC3339), rngStr)
	}
}

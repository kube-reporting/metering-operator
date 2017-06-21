package promsum

import (
	"testing"
	"time"
)

func TestBillingRecord_Range(t *testing.T) {
	record := BillingRecord{
		Labels: map[string]string{
			"test": "haldo",
		},
		Query:   "test query",
		Subject: "cluster4",
		Amount:  2323.22,
		Start:   time.Unix(343434, 0),
		End:     time.Unix(600000, 0),
	}

	rng := record.Range()
	if !rng.Start.Equal(record.Start) {
		t.Errorf("record start (%v) does not match range start (%v)", record.Start, rng.Start)
	}

	if !rng.End.Equal(record.End) {
		t.Errorf("record end (%v) does not match range end (%v)", record.End, rng.End)
	}
}

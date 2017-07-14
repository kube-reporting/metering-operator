package promsum

import (
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
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

func TestBillingRecord_Prorate(t *testing.T) {
	data := []struct {
		record BillingRecord
		rng    cb.Range
		amount float64
	}{
		{ // entire range
			record: BillingRecord{
				Amount: 100,
				Start:  time.Unix(100, 0),
				End:    time.Unix(200, 0),
			},
			rng: cb.Range{
				Start: time.Unix(100, 0),
				End:   time.Unix(200, 0),
			},
			amount: 100,
		},
		{ // half range
			record: BillingRecord{
				Amount: 100,
				Start:  time.Unix(100, 0),
				End:    time.Unix(200, 0),
			},
			rng: cb.Range{
				Start: time.Unix(150, 0),
				End:   time.Unix(200, 0),
			},
			amount: 50,
		},
	}

	for _, e := range data {
		prorate, err := e.record.Prorate(e.rng)
		if err != nil {
			t.Errorf("could not prorate record %v: %v", e.record, err)
		}

		if prorate.Amount != e.amount {
			t.Errorf("prorated amount unexpected: got %f, want %f", prorate.Amount, e.amount)
		}

		if prorate.Query != e.record.Query {
			t.Errorf("query unexpected: got %s, want %s", prorate.Query, e.record.Query)
		}

		if prorate.Subject != e.record.Subject {
			t.Errorf("subject unexpected: got %s, want %s", prorate.Subject, e.record.Subject)
		}
	}
}

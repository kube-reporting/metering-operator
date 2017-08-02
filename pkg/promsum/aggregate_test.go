package promsum

import (
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

var (
	// PodLabels is the lab***REMOVED***t that a Pod might have.
	PodLabels = map[string]string{
		"pod":       "k8s-prometheus-0",
		"namespace": "tectonic-system",
	}
)

func TestAggregate(t *testing.T) {
	now := time.Now().UTC()
	data := []struct {
		in          []BillingRecord
		rng         cb.Range
		mergeLabels []string
		out         []BillingRecord
		err         string
	}{ // entire range is inside one record
		{
			in: []BillingRecord{
				{
					Start:   now.Add(-20 * time.Minute),
					End:     now,
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  100,
				},
				{
					Start:   now,
					End:     now.Add(20 * time.Minute),
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  100,
				},
			},
			rng: cb.Range{
				Start: now,
				End:   now.Add(10 * time.Minute),
			},
			mergeLabels: []string{"pod", "namespace"},
			out: []BillingRecord{
				{
					Start:   now,
					End:     now.Add(10 * time.Minute),
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  50,
				},
			},
		},
		{
			in: []BillingRecord{
				{ // 25 from here
					Start:   now.Add(-20 * time.Minute),
					End:     now,
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  100,
				},
				{ // 100 fom here
					Start:   now,
					End:     now.Add(20 * time.Minute),
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  100,
				},
				{ // 100 from here
					Start:   now.Add(20 * time.Minute),
					End:     now.Add(40 * time.Minute),
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  200,
				},
			},
			rng: cb.Range{
				Start: now.Add(-5 * time.Minute),
				End:   now.Add(30 * time.Minute),
			},
			mergeLabels: []string{"pod", "namespace"},
			out: []BillingRecord{
				{
					Start:   now.Add(-5 * time.Minute),
					End:     now.Add(30 * time.Minute),
					Subject: "subject",
					Query:   "num_cats",
					Labels:  PodLabels,
					Amount:  225,
				},
			},
		},
	}

	for _, e := range data {
		actual, err := Aggregate(e.in, e.rng, e.mergeLabels)
		if err != nil {
			t.Errorf("could not create aggregate of %d records over %v: %v", len(e.in), e.rng, err)
		}

		if len(actual) != len(e.out) {
			t.Errorf("Unexpected number of records after aggregation: got %d, want %d", len(actual), len(e.out))
		}

		for _, actualRecord := range actual {
			for _, expectedRecord := range e.out {
				if actualRecord.Amount != expectedRecord.Amount {
					t.Errorf("unexpected amount after aggregation: got %f, want %f", actualRecord.Amount, expectedRecord.Amount)
				}

				if !actualRecord.Start.Equal(expectedRecord.Start) {
					t.Errorf("unexpected start after aggregation: got %v, want %v", actualRecord.Start, expectedRecord.Start)
				}

				if !actualRecord.End.Equal(expectedRecord.End) {
					t.Errorf("unexpected end after aggregation: got %v, want %v", actualRecord.End, expectedRecord.End)
				}
			}
		}

	}
}

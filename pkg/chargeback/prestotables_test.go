package chargeback

import (
	"reflect"
	"testing"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
)

func TestRemovePartition(t *testing.T) {
	type test struct {
		in       cbTypes.PrestoTable
		toRemove cbTypes.PrestoTablePartition
		out      cbTypes.PrestoTable
	}

	tests := []test{
		{
			in: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
					},
				},
			},
			toRemove: cbTypes.PrestoTablePartition{
				Start:    "20170201",
				End:      "20170301",
				Location: "foobar",
			},
			out: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
					},
				},
			},
		},
		{
			in: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
						{
							Start:    "20170201",
							End:      "20170301",
							Location: "foobar",
						},
						{
							Start:    "20170301",
							End:      "20170401",
							Location: "foobar",
						},
					},
				},
			},
			toRemove: cbTypes.PrestoTablePartition{
				Start:    "20170201",
				End:      "20170301",
				Location: "foobar",
			},
			out: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
						{
							Start:    "20170301",
							End:      "20170401",
							Location: "foobar",
						},
					},
				},
			},
		},
		{
			in: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
						{
							Start:    "20170201",
							End:      "20170301",
							Location: "foobar",
						},
						{
							Start:    "20170301",
							End:      "20170401",
							Location: "foobar",
						},
					},
				},
			},
			toRemove: cbTypes.PrestoTablePartition{
				Start:    "20170101",
				End:      "20170201",
				Location: "foobar",
			},
			out: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170201",
							End:      "20170301",
							Location: "foobar",
						},
						{
							Start:    "20170301",
							End:      "20170401",
							Location: "foobar",
						},
					},
				},
			},
		},
		{
			in: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
						{
							Start:    "20170201",
							End:      "20170301",
							Location: "foobar",
						},
						{
							Start:    "20170301",
							End:      "20170401",
							Location: "foobar",
						},
					},
				},
			},
			toRemove: cbTypes.PrestoTablePartition{
				Start:    "20170301",
				End:      "20170401",
				Location: "foobar",
			},
			out: cbTypes.PrestoTable{
				State: cbTypes.PrestoTableState{
					Partitions: []cbTypes.PrestoTablePartition{
						{
							Start:    "20170101",
							End:      "20170201",
							Location: "foobar",
						},
						{
							Start:    "20170201",
							End:      "20170301",
							Location: "foobar",
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		removePartition(test.toRemove, &test.in)
		if !reflect.DeepEqual(test.out, test.in) {
			t.Errorf("test #%d: results don't match expected output:\nexpected: %v\nactual: %v", i, test.out, test.in)
		}
	}
}

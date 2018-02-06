package chargeback

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
)

func TestGetPartitionChanges(t *testing.T) {
	tests := []struct {
		name             string
		current          []cbTypes.PrestoTablePartition
		desired          []cbTypes.PrestoTablePartition
		expectedToRemove []cbTypes.PrestoTablePartition
		expectedToAdd    []cbTypes.PrestoTablePartition
	}{
		{
			name: "current and desired are same",
			current: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
			desired: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
		},
		{
			name: "empty current should have desired in to add list",
			desired: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
			expectedToAdd: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
		},
		{
			name: "desired is empty and current should be removed",
			current: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
			expectedToRemove: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
		},
		{
			name: "desired matches and replaces current",
			current: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
			desired: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "***REMOVED***zbuzz",
				},
			},
			// Remove the old one because start and end match the new desired
			// one
			expectedToRemove: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "foobar",
				},
			},
			expectedToAdd: []cbTypes.PrestoTablePartition{
				cbTypes.PrestoTablePartition{
					Start:    "20170101",
					End:      "20170201",
					Location: "***REMOVED***zbuzz",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changes := getPartitionChanges(test.current, test.desired)
			assert.Equal(t, test.expectedToAdd, changes.toAddPartitions, "to add should match expected to add")
			assert.Equal(t, test.expectedToRemove, changes.toRemovePartitions, "to remove should match expected to remove")
		})
	}
}

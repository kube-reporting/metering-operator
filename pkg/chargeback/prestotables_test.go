package chargeback

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

func TestGetPartitionChanges(t *testing.T) {
	tests := []struct {
		name             string
		current          []cbTypes.TablePartition
		desired          []cbTypes.TablePartition
		expectedToRemove []cbTypes.TablePartition
		expectedToAdd    []cbTypes.TablePartition
		expectedToUpdate []cbTypes.TablePartition
	}{
		{
			name: "current and desired are same",
			current: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "empty current should have desired in to add list",
			desired: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToAdd: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "desired is empty and current should be removed",
			current: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToRemove: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "desired matches and replaces current",
			current: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "***REMOVED***zbuzz",
				},
			},
			expectedToUpdate: []cbTypes.TablePartition{
				{
					PartitionSpec: presto.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
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

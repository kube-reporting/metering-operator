package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

func TestGetPartitionChanges(t *testing.T) {
	tests := []struct {
		name             string
		current          []cbTypes.HiveTablePartition
		desired          []cbTypes.HiveTablePartition
		expectedToRemove []cbTypes.HiveTablePartition
		expectedToAdd    []cbTypes.HiveTablePartition
		expectedToUpdate []cbTypes.HiveTablePartition
	}{
		{
			name: "current and desired are same",
			current: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "empty current should have desired in to add list",
			desired: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToAdd: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "desired is empty and current should be removed",
			current: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToRemove: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
		},
		{
			name: "desired matches and replaces current",
			current: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "***REMOVED***zbuzz",
				},
			},
			expectedToUpdate: []cbTypes.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
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
			partitionColumns := []hive.Column{{Name: "start", Type: "string"}, {Name: "end", Type: "string"}}
			changes := getPartitionChanges(partitionColumns, test.current, test.desired)
			assert.Equal(t, test.expectedToAdd, changes.toAddPartitions, "to add should match expected to add")
			assert.Equal(t, test.expectedToRemove, changes.toRemovePartitions, "to remove should match expected to remove")
		})
	}
}

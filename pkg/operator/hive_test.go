package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

func TestGetPartitionChanges(t *testing.T) {
	tests := []struct {
		name             string
		current          []metering.HiveTablePartition
		desired          []metering.HiveTablePartition
		expectedToRemove []metering.HiveTablePartition
		expectedToAdd    []metering.HiveTablePartition
		expectedToUpdate []metering.HiveTablePartition
	}{
		{
			name: "current and desired are same",
			current: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []metering.HiveTablePartition{
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
			desired: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToAdd: []metering.HiveTablePartition{
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
			current: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			expectedToRemove: []metering.HiveTablePartition{
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
			current: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "foobar",
				},
			},
			desired: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "fizbuzz",
				},
			},
			expectedToUpdate: []metering.HiveTablePartition{
				{
					PartitionSpec: hive.PartitionSpec{
						"start": "20170101",
						"end":   "20170201",
					},
					Location: "fizbuzz",
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

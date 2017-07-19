package aws

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	// timestampFields are the AWS billing ***REMOVED***elds that should be created in Hive as timestamps.
	timestampFields = []string{
		"bill_billingperiodstartdate",
		"bill_billingperiodenddate",
		"lineitem_usagestartdate",
		"lineitem_usageenddate",
	}

	// doubleFields are created as the Hive double type.
	doubleFields = []string{
		"lineitem_blendedcost",
	}
)

// Column is a description of a ***REMOVED***eld from a AWS usage report manifest ***REMOVED***le.
type Column struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// HiveName is the identi***REMOVED***er used for Hive columns.
func (c Column) HiveName() string {
	name := fmt.Sprintf("%s_%s", c.Category, c.Name)
	// hive does not allow ':' or '.' in identi***REMOVED***ers
	name = strings.Replace(name, ":", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return strings.ToLower(name)
}

// HiveType is the data type a column is created as in Hive.
func (c Column) HiveType() string {
	for _, col := range timestampFields {
		if c.HiveName() == col {
			return "timestamp"
		}
	}

	for _, col := range doubleFields {
		if c.HiveName() == col {
			return "double"
		}
	}
	return "string"
}

// Columns are a set of AWS Usage columns.
type Columns []Column

// HQL returns the columns formatted for a HiveQL CREATE statement.
// Duplicate columns will be suf***REMOVED***xed by an incrementing ordinal. This can happen with user de***REMOVED***ned ***REMOVED***elds like tags.
func (cols Columns) HQL() []string {
	out := make([]string, len(cols))
	seen := make(map[string]int, len(cols))

	for i, c := range cols {
		name := c.HiveName()

		// prevent duplicates by numbering them
		times, exists := seen[name]
		if exists {
			name += strconv.Itoa(times)
		}
		seen[name] = times + 1

		out[i] = fmt.Sprintf("`%s` %s", name, c.HiveType())
	}
	return out
}

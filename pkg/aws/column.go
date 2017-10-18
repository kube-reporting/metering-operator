package aws

import (
	"fmt"
	"strconv"
	"strings"
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
	switch c.HiveName() {
	case "lineitem_usagestartdate", "lineitem_usageenddate":
		return "timestamp"
	case "lineitem_blendedcost":
		return "double"
	default:
		return "string"
	}
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

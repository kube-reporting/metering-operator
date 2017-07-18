package aws

import (
	"fmt"
	"strings"
)

// Column is a description of a field from a AWS usage report manifest file.
type Column struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// TODO: handle duplicate column names
// HiveName is the identifier used for Hive columns.
func (c Column) HiveName() string {
	name := fmt.Sprintf("`%s_%s`", c.Category, c.Name)
	// hive does not allow ':' as an identifier
	name = strings.Replace(name, ":", "_", -1)
	return strings.Replace(name, ".", "_", -1)
}

// HiveType is the data type a column is created as in Hive.
func (c Column) HiveType() string {
	for _, col := range timestampFields {
		if c.HiveName() == col {
			return "timestamp"
		}
	}
	return "string"
}

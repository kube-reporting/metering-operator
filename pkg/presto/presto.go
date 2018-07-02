package presto

import (
	"fmt"

	_ "github.com/prestodb/presto-go-client/presto"

	"github.com/operator-framework/operator-metering/pkg/db"
)

type Row map[string]interface{}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PartitionSpec map[string]string

type TablePartition struct {
	Location      string        `json:"location"`
	PartitionSpec PartitionSpec `json:"partitionSpec"`
}

func ExecuteQuery(queryer db.Queryer, query string) error {
	rows, err := queryer.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Must call rows.Next() in order for errors to be populated correctly
	// because Query() only submits the query, and doesn't handle
	// success/failure. Next() is the method which inspects the submitted
	// queries status and causes errors to get stored in the sql.Rows object.
	for rows.Next() {
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("presto SQL error: %v", err)
	}
	return nil
}

// ExecuteSelectQuery performs the query on the table target. It's expected
// target has the correct schema.
func ExecuteSelect(queryer db.Queryer, query string) ([]Row, error) {
	rows, err := queryer.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []Row
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, Row(m))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

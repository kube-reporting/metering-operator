package hive

import (
	"net/url"
	"path"

	"github.com/operator-framework/operator-metering/pkg/db"
)

type Column struct {
	Name string
	Type string
}

type TableParameters struct {
	Name         string   `json:"name"`
	Columns      []Column `json:"columns"`
	Partitions   []Column `json:"partitions"`
	IgnoreExists bool     `json:"ignoreExists"`
}

type TableProperties struct {
	Location        string            `json:"location"`
	SerdeFormat     string            `json:"serdeFormat"`
	Format          string            `json:"format"`
	SerdeProperties map[string]string `json:"serdeProperties"`
	External        bool              `json:"external"`
}

// ExecuteCreateS3Table creates a new table backed by the given S3 bucket/pre***REMOVED***x with
// the speci***REMOVED***ed columns
func ExecuteCreateS3Table(queryer db.Queryer, tableName, bucket, pre***REMOVED***x string, columns []Column, dropTable bool) (TableParameters, TableProperties, error) {
	path := path.Join(pre***REMOVED***x, tableName)
	location, err := S3Location(bucket, path)
	if err != nil {
		return TableParameters{}, TableProperties{}, err
	}

	params := TableParameters{
		Name:         tableName,
		Columns:      columns,
		Partitions:   nil,
		IgnoreExists: false,
	}
	properties := TableProperties{
		Location:        location,
		SerdeFormat:     "",
		Format:          "",
		SerdeProperties: nil,
		External:        false,
	}
	err = ExecuteCreateTable(queryer, params, properties, dropTable)
	return params, properties, err

}

func ExecuteCreateTable(queryer db.Queryer, params TableParameters, properties TableProperties, dropTable bool) error {
	if dropTable {
		err := ExecuteDropTable(queryer, params.Name, true)
		if err != nil {
			return err
		}
	}

	query := generateCreateTableSQL(params, properties)
	_, err := queryer.Query(query)
	return err
}

func ExecuteDropTable(queryer db.Queryer, tableName string, ignoreNotExists bool) error {
	query := generateDropTableSQL(tableName, ignoreNotExists, true)
	_, err := queryer.Query(query)
	return err
}

// s3Location returns the HDFS path based on an S3 bucket and pre***REMOVED***x.
func S3Location(bucket, pre***REMOVED***x string) (string, error) {
	bucket = path.Join(bucket, pre***REMOVED***x)
	// Ensure the bucket URL has a trailing slash
	if bucket[len(bucket)-1] != '/' {
		bucket = bucket + "/"
	}
	location := "s3a://" + bucket

	locationURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return locationURL.String(), nil
}

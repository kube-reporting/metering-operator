package hive

import (
	"net/url"
	"path"

	"github.com/operator-framework/operator-metering/pkg/db"
)

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TableParameters struct {
	Name         string   `json:"name"`
	Columns      []Column `json:"columns"`
	Partitions   []Column `json:"partitions,omitempty"`
	IgnoreExists bool     `json:"ignoreExists"`
}

type TableProperties struct {
	Location           string            `json:"location,omitempty"`
	SerdeFormat        string            `json:"serdeFormat,omitempty"`
	FileFormat         string            `json:"***REMOVED***leFormat,omitempty"`
	SerdeRowProperties map[string]string `json:"serdeRowProperties,omitempty"`
	External           bool              `json:"external,omitempty"`
}

func ExecuteCreateTable(queryer db.Queryer, params TableParameters, properties TableProperties) error {
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

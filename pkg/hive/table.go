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

type SortColumn struct {
	Name      string `json:"name"`
	Decending *bool  `json:"decending,omitempty"`
}

type TableParameters struct {
	Database      string       `json:"database,omitempty"`
	Name          string       `json:"name"`
	Columns       []Column     `json:"columns"`
	PartitionedBy []Column     `json:"partitionedBy,omitempty"`
	ClusteredBy   []string     `json:"clusteredBy,omitempty"`
	SortedBy      []SortColumn `json:"sortedBy,omitempty"`
	NumBuckets    int          `json:"numBuckets,omitempty"`

	Location        string            `json:"location,omitempty"`
	RowFormat       string            `json:"rowFormat,omitempty"`
	FileFormat      string            `json:"fileFormat,omitempty"`
	TableProperties map[string]string `json:"tableProperties,omitempty"`
	External        bool              `json:"external,omitempty"`
}

type TablePartition struct {
	Location      string        `json:"location"`
	PartitionSpec PartitionSpec `json:"partitionSpec"`
}

type PartitionSpec map[string]string

func ExecuteCreateTable(queryer db.Queryer, params TableParameters, ignoreExists bool) error {
	query := generateCreateTableSQL(params, ignoreExists)
	_, err := queryer.Query(query)
	return err
}

func ExecuteDropTable(queryer db.Queryer, dbName, tableName string, ignoreNotExists bool) error {
	query := generateDropTableSQL(dbName, tableName, ignoreNotExists, false)
	_, err := queryer.Query(query)
	return err
}

// s3Location returns the HDFS path based on an S3 bucket and prefix.
func S3Location(bucket, prefix string) (string, error) {
	bucket = path.Join(bucket, prefix)
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

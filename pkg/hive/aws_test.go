package hive

import (
	"fmt"
	"testing"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

func TestCreateAWSUsageTable(t *testing.T) {
	hiveHost, s3Bucket, _ := setupHiveTest(t)

	conn, err := Connect(hiveHost)
	if err != nil {
		t.Fatal("error connecting: ", err)
	}
	defer conn.Close()

	dropQuery := fmt.Sprint("DROP TABLE ", AWSUsageTableName)
	if err = conn.Query(dropQuery); err != nil {
		t.Errorf("Could not delete existing table: %v", err)
	}

	manifests := getAWSManifests(t)
	if err = CreateAWSUsageTable(conn, s3Bucket, manifests[0]); err != nil {
		t.Error("error perfoming query: ", err)
	}

	selectQuery := fmt.Sprint("SELECT * FROM ", AWSUsageTableName, " LIMIT 10")
	if err = conn.Query(selectQuery); err != nil {
		t.Error("could not select from sample data: ", err)
	}
}

func getAWSManifests(t *testing.T) []aws.Manifest {
	bucket, reportName := "coreos-team-chargeback", "team-chargeback-testing"
	reportPrefix := "coreos-detailed-billing/coreosinc//coreos-detailed-billing-001"
	begin := time.Date(2017, time.June, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2017, time.June, 29, 0, 0, 0, 0, time.UTC)
	rng := cb.Range{begin, end}
	manifests, err := aws.RetrieveManifests(bucket, reportPrefix, reportName, rng)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	return manifests
}

package presto

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func TestRunAWSPodDollarReport(t *testing.T) {
	prestoHost := setupPrestoTest(t)
	connStr := fmt.Sprintf("presto://%s/hive/default", prestoHost)
	db, err := sql.Open("prestgo", connStr)
	if err != nil {
		t.Fatalf("failed to connect to presto: %v", err)
	}
	defer db.Close()

	outTable := "billingReport1"
	begin := time.Date(2017, time.July, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2017, time.July, 29, 0, 0, 0, 0, time.UTC)
	rng := cb.Range{begin, end}
	if err = RunAWSPodDollarReport(db, hive.PromsumTableName, hive.AWSUsageTableName, outTable, rng); err != nil {
		t.Fatal("could not create report table ", err)
	}

	selectQuery := fmt.Sprint("SELECT COUNT(*) FROM ", outTable)
	if _, err = db.Query(selectQuery); err != nil {
		t.Errorf("Failed to load newly created table: %v", err)
	}
}

package hive

import (
	"fmt"
	"testing"
)

func TestCreateResultTable(t *testing.T) {
	hiveHost, s3Bucket, s3Prefix := setupHiveTest(t)

	conn, err := Connect(hiveHost)
	if err != nil {
		t.Fatal("error connecting: ", err)
	}
	defer conn.Close()

	resultTable := "billingReport1"
	s3Prefix = resultTable + "/"
	dropQuery := fmt.Sprint("DROP TABLE ", resultTable)
	if err = conn.Query(dropQuery); err != nil {
		t.Errorf("Could not delete existing table: %v", err)
	}

	if err = CreatePodCostTable(conn, resultTable, s3Bucket, s3Prefix); err != nil {
		t.Error("error perfoming query: ", err)
	}

	selectQuery := fmt.Sprint("SELECT * FROM ", resultTable)
	if err = conn.Query(selectQuery); err != nil {
		t.Error("could not select from sample data: ", err)
	}
}

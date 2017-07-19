package hive

import (
	"fmt"
	"testing"
)

func TestCreatePromsumTable(t *testing.T) {
	hiveHost, s3Bucket, s3Pre***REMOVED***x := setupHiveTest(t)

	conn, err := Connect(hiveHost)
	if err != nil {
		t.Fatal("error connecting: ", err)
	}
	defer conn.Close()

	dropQuery := fmt.Sprint("DROP TABLE ", PromsumTableName)
	if err = conn.Query(dropQuery); err != nil {
		t.Errorf("Could not delete existing table: %v", err)
	}

	if err = CreatePromsumTable(conn, PromsumTableName, s3Bucket, s3Pre***REMOVED***x); err != nil {
		t.Error("error perfoming query: ", err)
	}

	selectQuery := fmt.Sprint("SELECT * FROM ", PromsumTableName)
	if err = conn.Query(selectQuery); err != nil {
		t.Error("could not select from sample data: ", err)
	}
}

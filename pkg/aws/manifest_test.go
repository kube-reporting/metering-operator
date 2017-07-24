package aws

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

const (
	manifestText = `{
  "assemblyId":"ea74f90b-e82f-9c72-fab6-abc716793752",
  "account":"826591639284",
  "columns":[{
    "category":"identity",
    "name":"LineItemId"
  },{
    "category":"identity",
    "name":"TimeInterval"
  },{
    "category":"bill",
    "name":"InvoiceId"
  },{
    "category":"bill",
    "name":"BillingEntity"
  },{
    "category":"bill",
    "name":"BillType"
  }],
  "charset":"UTF-8",
  "compression":"GZIP",
  "contentType":"text/csv",
  "reportId":"494124bac4e25a16a3b704c13be2c525fd60d25b0675eb0f72e7b9e8ea09e167",
  "reportName":"sample-report",
  "billingPeriod":{
    "start":"20170701T000000.000Z",
    "end":"20170801T000000.000Z"
  },
  "bucket":"billing-bucket",
  "reportKeys":["billing-path/20170701-20170801/ea74f90b-e82f-9c72-fab6-abc716793752/sample-report-1.csv.gz","billing-path/20170701-20170801/ea74f90b-e82f-9c72-fab6-abc716793752/sample-report-2.csv.gz"],
  "additionalArtifactKeys":[]
}`
)

func TestManifest_Paths(t *testing.T) {
	var manifest Manifest
	if err := json.Unmarshal([]byte(manifestText), &manifest); err != nil {
		t.Error("failed to marshal error: ", err)
	}

	numExpectedPaths := 1
	expectedPath := "billing-path/20170701-20170801/ea74f90b-e82f-9c72-fab6-abc716793752"
	if paths := manifest.Paths(); len(paths) != numExpectedPaths {
		t.Errorf("unexpected number of paths: got %d, want %d", len(paths), numExpectedPaths)
	} ***REMOVED*** if paths[0] != expectedPath {
		t.Errorf("unexpected path: got %s, want %s", paths[0], expectedPath)
	}

	// manifests without report keys should not return paths
	manifest.ReportKeys = nil
	if paths := manifest.Paths(); len(paths) != 0 {
		t.Error("manifests without report keys should not produce paths")
	}
}

func TestRetrieveManifests(t *testing.T) {
	bucket, reportName := "coreos-team-chargeback", "team-chargeback-testing"
	reportPre***REMOVED***x := "coreos-detailed-billing/coreosinc//coreos-detailed-billing-001"
	begin := time.Date(2017, time.June, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2017, time.June, 29, 0, 0, 0, 0, time.UTC)
	rng := cb.Range{begin, end}
	manifests, err := RetrieveManifests(bucket, reportPre***REMOVED***x, reportName, rng)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	for _, m := range manifests {
		fmt.Println("Start: ", m.BillingPeriod.Start, ", End: ", m.BillingPeriod.End)
	}
}

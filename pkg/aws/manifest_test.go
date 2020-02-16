package aws

import (
	"encoding/json"
	"testing"
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
	if paths := manifest.paths(); len(paths) != numExpectedPaths {
		t.Errorf("unexpected number of paths: got %d, want %d", len(paths), numExpectedPaths)
	} else if paths[0] != expectedPath {
		t.Errorf("unexpected path: got %s, want %s", paths[0], expectedPath)
	}

	// manifests without report keys should not return paths
	manifest.ReportKeys = nil
	if paths := manifest.paths(); len(paths) != 0 {
		t.Error("manifests without report keys should not produce paths")
	}
}

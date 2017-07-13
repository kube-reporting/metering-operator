package aws

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalManifest(t *testing.T) {
	manifestText := `{
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

	var manifest Manifest
	if err := json.Unmarshal([]byte(manifestText), &manifest); err != nil {
		t.Error("failed to marshal error: ", err)
	}
}

package chargeback

import (
	"encoding/json"
	"testing"
)

func TestReportPhase(t *testing.T) {
	badReport := []byte(`{"status":{"phase":"closed"}}`)

	var actualReport Report
	if err := json.Unmarshal(badReport, &actualReport); err == nil {
		t.Error("should have errored, bad report")
	}

	goodReport := []byte(`{"status":{"phase":"Started"}}`)
	if err := json.Unmarshal(goodReport, &actualReport); err != nil {
		t.Error("shouldn't have errored: ", err)
	} else if actualReport.Status.Phase != ReportPhaseStarted {
		t.Errorf("phase mismatch: want %s, got %s", ReportPhaseStarted, actualReport.Status.Phase)
	}
}

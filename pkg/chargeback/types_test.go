package chargeback

import (
	"encoding/json"
	"testing"
)

func TestQueryPhase(t *testing.T) {
	badQuery := []byte(`{"status":{"phase":"closed"}}`)

	var actualQuery Query
	if err := json.Unmarshal(badQuery, &actualQuery); err == nil {
		t.Error("should have errored, bad query")
	}

	goodQuery := []byte(`{"status":{"phase":"Started"}}`)
	if err := json.Unmarshal(goodQuery, &actualQuery); err != nil {
		t.Error("shouldn't have errored: ", err)
	} else if actualQuery.Status.Phase != QueryPhaseStarted {
		t.Errorf("phase mismatch: want %s, got %s", QueryPhaseStarted, actualQuery.Status.Phase)
	}
}

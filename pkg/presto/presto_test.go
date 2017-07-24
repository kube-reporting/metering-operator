package presto

import (
	"os"
	"testing"
)

var (
	// PrestoHostVar is environment variable holding the Presto host used for testing.
	PrestoHostVar = "TEST_PRESTO_HOST"
)

func setupPrestoTest(t *testing.T) (prestoHost string) {
	var exists bool
	if prestoHost, exists = os.LookupEnv(PrestoHostVar); !exists {
		t.Skipf("To test Presto, set the '%s' to the Presto instance to be used.", PrestoHostVar)
		t.SkipNow()
	}
	return
}

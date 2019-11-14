package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/test/reportingframework"
)

func testInvalidMeteringCon***REMOVED***gMissingStorageSpec(t *testing.T, rf *reportingframework.ReportingFramework) {
	require.NotNil(t, rf, "expected the reportingframework object would not be nil")
	require.NotNil(t, rf.MeteringClient, "expected the reportingframework.MeteringClient ***REMOVED***eld would not be nil")
	require.NotEmpty(t, rf.Namespace, "expected the reportingframework.Namespace ***REMOVED***eld would not be empty")

	mc, err := rf.MeteringClient.MeteringCon***REMOVED***gs(rf.Namespace).Get("operator-metering", meta.GetOptions{})
	require.Truef(t, apierrors.IsNotFound(err), "expected the MeteringCon***REMOVED***g to not exist, got: %v, err: %v", mc, err)
}

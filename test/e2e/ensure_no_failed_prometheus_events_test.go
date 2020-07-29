package e2e

import (
	"context"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const failedPrometheusQueryEventReason = "FailedPrometheusQuery"

func testFailedPrometheusQueryEvents(t *testing.T, rf *reportingframework.ReportingFramework) {
	events, err := rf.KubeClient.CoreV1().Events(rf.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err, "failed to list the events in the test namespace")

	var found bool
	for _, event := range events.Items {
		if event.Reason == failedPrometheusQueryEventReason {
			found = true
			break
		}
	}

	assert.False(t, found, "expected to find no events for failed Prometheus queries")
}

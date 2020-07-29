package e2e

import (
	"context"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const leaderElectionName = "reporting-operator-leader-lease"

func testLeaderElectionEventIsCreated(t *testing.T, rf *reportingframework.ReportingFramework) {
	events, err := rf.KubeClient.CoreV1().Events(rf.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err, "failed to list the events in the test namespace")

	var found bool
	for _, event := range events.Items {
		if event.InvolvedObject.Kind != "ConfigMap" {
			continue
		}
		if event.InvolvedObject.Name == leaderElectionName {
			found = true
		}
	}

	assert.True(t, found, "expected to find the leader election event created")
}

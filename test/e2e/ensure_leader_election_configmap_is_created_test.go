package e2e

import (
	"context"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const leaderElectionConfigMapName = "metering-operator"

// testLeaderElctionConfigMapIsCreated is reponsible for querying for the metering operator
// leader election lock ConfigMap. Note: this implementation is subject to change based on the
// ansible-operator v1 migration. Before, a ConfigMap is created using the `POD_NAME` environment
// variable exposed using the k8s downward API. Now, leader election is determined using
// controller runtime's implementation.
func testLeaderElectionConfigMapIsCreated(t *testing.T, rf *reportingframework.ReportingFramework) {
	_, err := rf.KubeClient.CoreV1().ConfigMaps(rf.Namespace).Get(context.Background(), leaderElectionConfigMapName, metav1.GetOptions{})
	require.NoError(t, err, "failed to query for the %s leader election ConfigMap", leaderElectionConfigMapName)
}

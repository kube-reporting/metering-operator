package e2e

import (
	"context"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: change this name once metering has migrated to using
// the ansible-operator v1.x version.
const leaderElectionConfigMapName = "metering-operator-lock"

// testLeaderElctionConfigMapIsCreated is reponsible for querying for the metering operator
// leader election lock ConfigMap. Note: this implementation is subject to change based on the
// ansible-operator v1 migration. Currently, a ConfigMap is created using the `POD_NAME` environment
// variable exposed using the k8s downward API. In the ansible-operator 1.x world, leader election
// is determined using controller runtime's implementation.
func testLeaderElectionConfigMapIsCreated(t *testing.T, rf *reportingframework.ReportingFramework) {
	_, err := rf.KubeClient.CoreV1().ConfigMaps(rf.Namespace).Get(context.Background(), leaderElectionConfigMapName, metav1.GetOptions{})
	require.NoError(t, err, "failed to query for the %s leader election ConfigMap", leaderElectionConfigMapName)
}

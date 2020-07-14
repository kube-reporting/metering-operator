package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testNodeSelectorConfigurationWorks(t *testing.T, rf *reportingframework.ReportingFramework) {
	// note: we already know the node selector configuration that
	// was specified in the MeteringConfig YAML manifest.
	expectedNodeSelector := map[string]string{
		"node-role.kubernetes.io/worker": "",
		"metering-node-testing-label":    "true",
	}

	tt := []struct {
		Name                 string
		PodNameContains      string
		ExpectedNodeSelector map[string]string
	}{
		{
			Name:                 "valid-presto-coordinator-node-selector",
			PodNameContains:      "presto-coordinator-0",
			ExpectedNodeSelector: expectedNodeSelector,
		},
		{
			Name:                 "valid-hive-server-node-selector",
			PodNameContains:      "hive-server-0",
			ExpectedNodeSelector: expectedNodeSelector,
		},
		{
			Name:                 "valid-hive-metastore-node-selector",
			PodNameContains:      "hive-metastore-0",
			ExpectedNodeSelector: expectedNodeSelector,
		},
		{
			Name:                 "valid-hdfs-datanode-node-selector",
			PodNameContains:      "hdfs-datanode-0",
			ExpectedNodeSelector: expectedNodeSelector,
		},
		{
			Name:                 "valid-hdfs-namenode-node-selector",
			PodNameContains:      "hdfs-namenode-0",
			ExpectedNodeSelector: expectedNodeSelector,
		},
		{
			Name:                 "valid-reporting-operator-node-selector",
			PodNameContains:      "reporting-operator",
			ExpectedNodeSelector: expectedNodeSelector,
		},
	}

	pods, err := rf.KubeClient.CoreV1().Pods(rf.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err, "expected querying the %s namespace for the list of pods would produce no error", rf.Namespace)

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var matched bool
			for _, pod := range pods.Items {
				// in the case where a pod was spun up by a deployment
				// controller, check if the current pod were handling
				// in this loop iteration contains a subset of the test
				// case pod name.
				if !strings.Contains(pod.Name, tc.PodNameContains) {
					continue
				}

				assert.Equal(t, tc.ExpectedNodeSelector, pod.Spec.NodeSelector, "expected that the actual node selectors for the Pod would matched the test case expected value")
				matched = true
			}
			assert.True(t, matched, "expected to find the pod listed in the test case")
		})
	}
}

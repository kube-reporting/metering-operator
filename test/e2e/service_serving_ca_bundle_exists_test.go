package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

func testReportingOperatorServiceCABundleExists(t *testing.T, rf *reportingframework.ReportingFramework) {
	// Attempt to grab the reporting-operator Pod in the rf.Namespace.
	// Because this Pod is spun up using a Deployment instead of a Statefulset
	// we need to list any Pods matching the app=reporting-operator label as we
	// don't know the name ahead of time.
	podList, err := rf.KubeClient.CoreV1().Pods(rf.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: "app=reporting-operator",
	})
	require.NoError(t, err, "failed to list the reporting-operator Pod in the %s namespace", rf.Namespace)
	require.Len(t, podList.Items, 1, "expected the list of app=reporting-operator Pods in the %s namespace to match a length of 1", rf.Namespace)

	// We expect there's only going to be one reporting-operator
	// in a single namespace, so hardcode the reporting-operator Pod
	// object to be the first item in the list returned.
	pod := podList.Items[0]
	// Start building up a command that we'll execute against the
	// reporting-operator container to ensure that the service-ca-bundle
	// exists in the correct location.
	execCmd := "ls /var/run/configmaps/service-ca-bundle"

	for _, container := range pod.Spec.Containers {
		t.Run(container.Name, func(t *testing.T) {
			t.Parallel()

			options := testhelpers.NewExecOptions(pod.Name, pod.Namespace, container.Name, false, strings.Fields(execCmd))
			stdoutBuf, stderrBuf, err := testhelpers.ExecPodCommandWithOptions(rf.KubeConfig, rf.KubeClient, options)
			require.NoError(t, err, "failed to successfully exec into the %s Pod in the %s namespace", pod.Name, pod.Namespace)

			assert.Contains(t, stdoutBuf.String(), "service-ca.crt", "expected that the service-ca.crt would be present in the reporting-operator container")
			assert.Len(t, stderrBuf.String(), 0, "expected the stderr buffer would be nil")
		})
	}
}

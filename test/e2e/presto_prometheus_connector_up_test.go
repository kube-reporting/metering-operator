package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

const (
	prestoCoordinatorPodName = "presto-coordinator-0"
	prestoContainerName      = "presto"
)

func testPrometheusConnectorWorks(t *testing.T, rf *reportingframework.ReportingFramework) {
	pod, err := rf.KubeClient.CoreV1().Pods(rf.Namespace).Get(context.Background(), prestoCoordinatorPodName, metav1.GetOptions{})
	require.Nil(t, err, "expected there would be no error querying for the presto-coordinator-0 Pod in the %s namespace", rf.Namespace)

	t.Logf("Found the %s pod in the %s namespace", pod.Name, pod.Namespace)

	execCmd := []string{
		"presto-cli",
		"--server",
		"https://presto:8080",
		"--user",
		"root",
		"--catalog",
		"prometheus",
		"--schema",
		"default",
		"--keystore-path",
		"/opt/presto/tls/keystore.pem",
		"--execute",
		"SELECT * FROM prometheus.default.up where timestamp > (NOW() - INTERVAL '10' second)",
	}
	options := testhelpers.NewExecOptions(pod.Name, pod.Namespace, prestoContainerName, false, execCmd)

	stdoutBuf, stderrBuf, err := testhelpers.ExecPodCommandWithOptions(rf.KubeConfig, rf.KubeClient, options)
	require.Nil(t, err, "expected running the a presto query would produce no error")
	require.Containsf(t, stdoutBuf.String(), "endpoint=metrics", "expected the presto query output would contain `endpoint=metrics`")
	require.Len(t, stderrBuf.String(), 0, "expected that the stderr buffer would return nothing")
}

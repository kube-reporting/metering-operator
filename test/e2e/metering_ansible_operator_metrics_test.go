package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

const (
	prometheusPodName       = "presto-k8s-0"
	prometheusNamespace     = "openshift-monitoring"
	prometheusContainerName = "prometheus"
	prometheusQueryEndpoint = "http://127.0.0.1:9090/api/v1/query"
)

// testMeteringAnsibleOperatorMetricsWork ensures that the ServiceMonitor object the ansible-operator
// creates for the metering-ansible-operator is exposing metrics correctly. It execs in the prometheus-k8s-0
// Pod in the openshift-monitoring namespace and fires off a known metric that gets collected in the test
// namespace. This is a follow-up to the fixes made in https://github.com/kube-reporting/metering-operator/pull/1235.
func testMeteringAnsibleOperatorMetricsWork(t *testing.T, rf *reportingframework.ReportingFramework) {
	query := fmt.Sprintf("query=workqueue_work_duration_seconds_count{namespace='%s',job='metering-operator-metrics'}", rf.Namespace)

	cmd := []string{
		"curl", "-k", "-G", "-s", "--data-urlencode", query, prometheusQueryEndpoint,
	}
	options := testhelpers.NewExecOptions(prometheusPodName, prometheusNamespace, prometheusContainerName, false, cmd)
	stdoutBuf, stderrBuf, err := testhelpers.ExecPodCommandWithOptions(rf.KubeConfig, rf.KubeClient, options)
	t.Logf("Stdout output: %s", stdoutBuf.String())
	t.Logf("Stderr output: %s", stderrBuf.String())
	require.Nil(t, err, "expected running the prometheus query would produce no error")
	require.Containsf(t, stdoutBuf.String(), "workqueue_work_duration_seconds_count", "expected the prometheus query would return the 'workqueue_work_duration_seconds_count' job name")
	require.Containsf(t, stdoutBuf.String(), "success", "expected the prometheus query would return a successful status")
	require.Len(t, stderrBuf.String(), 0, "expected that the stderr buffer would return nothing")
}

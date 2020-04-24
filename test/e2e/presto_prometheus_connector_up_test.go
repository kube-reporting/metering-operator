package e2e

import (
	"bytes"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"strings"

	"testing"
)

func testPrometheusConnectorWorks(t *testing.T, rf *reportingframework.ReportingFramework) {
	t.Helper()

	// query for the presto-coordinator pod name
	podNameCmd := exec.Command(
		"kubectl",
		"-n", rf.Namespace,
		"get", "pods",
		"-l", "app=presto,presto=coordinator",
		"-o", "name",
	)

	var prestoHostResults bytes.Buffer
	podNameCmd.Stderr = os.Stderr
	podNameCmd.Stdout = &prestoHostResults

	err := podNameCmd.Run()
	require.Nil(t, err, "expected querying for the presto pod name would produce no error")
	require.NotEmpty(t, prestoHostResults.String(), "unable to parse the presto-coordinator pod name")

	t.Logf("querying for the pod/presto-coordinator name returned: %s", prestoHostResults.String())

	// we know the output is going to be of the form `pod/<presto coordinator pod name>` so split by '/'
	tmp := strings.Split(prestoHostResults.String(), "/")[1]
	host := strings.TrimSuffix(tmp, "\n")
	t.Logf("host found for presto from prep phase: %s", host)

	// TODO: need to handle the case where TLS is disabled
	queryCmd := exec.Command(
		"kubectl",
		"-n", rf.Namespace,
		"exec",
		host,
		"-c", "presto",
		"--",
		"presto-cli",
		"--server", "https://presto:8080",
		"--user", "root",
		"--catalog", "prometheus",
		"--schema", "default",
		"--keystore-path", "/opt/presto/tls/keystore.pem",
		"--execute",
		"SELECT * FROM prometheus.default.up where timestamp > (NOW() - INTERVAL '10' second)",
	)

	var queryResults bytes.Buffer
	queryCmd.Stderr = os.Stderr
	queryCmd.Stdout = &queryResults

	err = queryCmd.Run()
	require.Nil(t, err, "expected running the a presto query would produce no error")
	require.Containsf(t, queryResults.String(), "endpoint=metrics", "expected the presto query output would contain `endpoint=metrics`")
}

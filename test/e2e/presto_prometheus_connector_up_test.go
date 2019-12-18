package e2e

import (
	"bytes"
	"github.com/operator-framework/operator-metering/test/reportingframework"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"strings"

	"testing"
)

func testPrometheusConnectorWorks(t *testing.T, rf *reportingframework.ReportingFramework) {
	t.Helper()
	cmdPrep := exec.Command(
		"kubectl",
		"-n", rf.Namespace,
		"get", "pods",
		"-l", "app=presto,presto=coordinator",
		"-o", "name",
	)
	var prestoHostResults bytes.Buffer
	cmdPrep.Stderr = os.Stderr
	cmdPrep.Stdout = &prestoHostResults
	err := cmdPrep.Run()
	t.Logf("pod name returned as: %s", string(prestoHostResults.Bytes()))
	require.Nil(t, err, "expected querying for the presto pod name would produce no error")
	require.NotEmpty(t, string(prestoHostResults.Bytes()), "unable to parse presto pod name")

	// we know the output is going to be of the form `pod/<presto coordinator pod name>` so split by '/'
	tmp := strings.Split(string(prestoHostResults.Bytes()), "/")[1]
	host := strings.TrimSuffix(tmp, "\n")
	t.Logf("host found for presto from prep phase: %s", host)

	cmdTest := exec.Command(
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
	cmdTest.Stderr = os.Stderr
	cmdTest.Stdout = &queryResults
	err = cmdTest.Run()
	require.Containsf(t, string(queryResults.Bytes()), "endpoint=metrics", "query output didn't contain `endpoint=metrics`")
}

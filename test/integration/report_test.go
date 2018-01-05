package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	chargebackv1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	reportTestTimeout = 5 * time.Minute
)

func init() {
	var err error
	if reportTestTimeoutStr := os.Getenv("REPORT_TEST_TIMEOUT"); reportTestTimeoutStr != "" {
		reportTestTimeout, err = time.ParseDuration(reportTestTimeoutStr)
		if err != nil {
			log.Fatalf("Invalid REPORT_TEST_TIMEOUT: %v", err)
		}
	}
}

func newSimpleReport(name, namespace, queryName string, start, end time.Time) *chargebackv1alpha1.Report {
	return &chargebackv1alpha1.Report{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: chargebackv1alpha1.ReportSpec{
			ReportingStart:      meta.Time{start},
			ReportingEnd:        meta.Time{end},
			GenerationQueryName: queryName,
			RunImmediately:      true,
		},
	}
}

func TestReportsProduceData(t *testing.T) {
	// reportEnd is 1 minute ago because Prometheus may not have collected
	// the most recent 1 minute of data yet
	reportEnd := time.Now().Add(-time.Minute)
	// To make things faster, let's limit the window to 10 minutes
	reportStart := reportEnd.Add(-10 * time.Minute)

	tests := []struct {
		// name is the name of the sub test but also the name of the report.
		name      string
		queryName string
		timeout   time.Duration
	}{
		{
			name:      "pod-cpu-usage-by-namespace",
			queryName: "pod-cpu-usage-by-namespace",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-usage-by-namespace",
			queryName: "pod-memory-usage-by-namespace",
			timeout:   reportTestTimeout + time.Minute,
		},
		{
			name:      "pod-cpu-usage-by-node",
			queryName: "pod-cpu-usage-by-node",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-usage-by-node",
			queryName: "pod-memory-usage-by-node",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-usage-by-node-with-usage-percent",
			queryName: "pod-memory-usage-by-node-with-usage-percent",
			timeout:   reportTestTimeout + time.Minute,
		},
		{
			name:      "node-cpu-usage",
			queryName: "node-cpu-usage",
			timeout:   reportTestTimeout,
		},
		{
			name:      "node-memory-usage",
			queryName: "node-memory-usage",
			timeout:   reportTestTimeout,
		},
		// TODO(chancez): Add AWS Reports
	}

	reqParams := chargeback.CollectPromsumDataRequest{
		StartTime: reportStart,
		EndTime:   reportEnd,
	}
	body, err := json.Marshal(reqParams)
	require.NoError(t, err, "should be able to json encode request parameters")
	req := testFramework.NewChargebackSVCPOSTRequest(testFramework.Namespace, "chargeback", "/api/v1/collect/prometheus", body)
	result := req.Do()
	resp, err := result.Raw()
	require.NoErrorf(t, err, "expected no errors triggering data collection, body: %v", string(resp))

	for i, test := range tests {
		// Fix closure captures
		test := test
		i := i
		// The JVM has a warm up time and the ***REMOVED***rst report always takes longer
		// than others, so give it a longer timeout
		if i == 0 {
			test.timeout += time.Minute
		}

		t.Run(test.name, func(t *testing.T) {
			if testing.Short() && i != 0 {
				t.Skipf("skipping test in short mode")
				return
			}

			report := newSimpleReport(test.name, testFramework.Namespace, test.queryName, reportStart, reportEnd)

			err := testFramework.ChargebackClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
			assert.Condition(t, func() bool {
				return err == nil || errors.IsNotFound(err)
			}, "failed to ensure report doesn't exist before creating report")

			t.Logf("creating report %s", report.Name)
			err = testFramework.CreateChargebackReport(testFramework.Namespace, report)
			require.NoError(t, err, "creating report should succeed")

			defer func() {
				t.Logf("deleting report %s", report.Name)
				err := testFramework.ChargebackClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete report to succeed")
			}()

			query := map[string]string{
				"name":   test.name,
				"format": "json",
			}

			var reportResults []map[string]interface{}
			err = wait.Poll(time.Second*5, test.timeout, func() (bool, error) {
				req := testFramework.NewChargebackSVCRequest(testFramework.Namespace, "chargeback", "/api/v1/reports/get", query)
				result := req.Do()
				resp, err := result.Raw()
				if err != nil {
					return false, fmt.Errorf("error querying chargeback service got error: %v, body: %v", err, string(resp))
				}

				statusCode := new(int)
				result.StatusCode(statusCode)

				if *statusCode == http.StatusAccepted {
					t.Logf("report is still running")
					return false, nil
				}

				require.Equal(t, http.StatusOK, *statusCode, "http response status code should be ok")

				err = json.Unmarshal(resp, &reportResults)
				require.NoError(t, err, "expected to unmarshal response")
				return true, nil
			})
			require.NoError(t, err, "expected getting report result to not timeout")
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")
		})
	}
}

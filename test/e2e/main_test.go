package e2e

import (
	"encoding/json"
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/operator-framework/operator-metering/pkg/chargeback"
	"github.com/operator-framework/operator-metering/test/framework"
)

var (
	testFramework     *framework.Framework
	collectOnce       sync.Once
	globalReportStart time.Time
	globalReportEnd   time.Time
)

func TestMain(m *testing.M) {
	kubecon***REMOVED***g := flag.String("kubecon***REMOVED***g", "", "kube con***REMOVED***g path, e.g. $HOME/.kube/con***REMOVED***g")
	ns := flag.String("namespace", "chargeback-ci", "test namespace")
	flag.Parse()

	var err error

	if testFramework, err = framework.New(*ns, *kubecon***REMOVED***g); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}

func collectMetricsOnce(t *testing.T) (reportStart time.Time, reportEnd time.Time) {
	t.Helper()
	collectOnce.Do(func() {
		// Use UTC, Prometheus uses UTCf for timestamps
		currentTime := time.Now().UTC()
		// reportEnd is an hour and 10 minutes ago because Prometheus may not
		// have collected very recent data, and setting to hour ago ensures
		// that a scheduledReport created with a LastReportTime of reportEnd
		// will run immediately.
		reportEnd = currentTime.Add(-(time.Hour + 10*time.Minute))
		// To make things faster, let's limit the window of collected data to
		// 10 minutes
		reportStart = reportEnd.Add(-10 * time.Minute)

		// stored so that future calls can immediately return the same
		// reportStart/reportEnd from above.
		globalReportEnd = reportEnd
		globalReportStart = reportStart

		reqParams := chargeback.CollectPromsumDataRequest{
			StartTime: reportStart,
			EndTime:   reportEnd,
		}
		body, err := json.Marshal(reqParams)
		require.NoError(t, err, "should be able to json encode request parameters")
		collectEndpoint := "/api/v1/datasources/prometheus/collect"
		t.Logf("Querying %s, currentTime: %s", collectEndpoint, currentTime)
		req := testFramework.NewChargebackSVCPOSTRequest(collectEndpoint, body)
		result := req.Do()
		resp, err := result.Raw()
		t.Logf("Finishing querying %s, took: %s to ***REMOVED***nish", collectEndpoint, time.Now().UTC().Sub(currentTime))
		require.NoErrorf(t, err, "expected no errors triggering data collection, body: %v", string(resp))
	})
	return globalReportStart, globalReportEnd
}

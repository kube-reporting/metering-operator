package framework

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/operator-framework/operator-metering/pkg/chargeback"
)

func (f *Framework) CollectMetricsOnce(t *testing.T) (time.Time, time.Time) {
	t.Helper()
	f.collectOnce.Do(func() {
		// Use UTC, Prometheus uses UTCf for timestamps
		currentTime := time.Now().UTC()

		// store the start/end so that future calls can immediately return the
		// same reportStart/reportEnd.

		// reportEnd is an hour and 10 minutes ago because Prometheus may not
		// have collected very recent data, and setting to hour ago ensures
		// that a scheduledReport created with a LastReportTime of reportEnd
		// will run immediately.
		f.reportEnd = currentTime.Add(-(time.Hour + 10*time.Minute))
		// To make things faster, let's limit the window of collected data to
		// 10 minutes
		f.reportStart = f.reportEnd.Add(-10 * time.Minute)

		reqParams := chargeback.CollectPromsumDataRequest{
			StartTime: f.reportStart,
			EndTime:   f.reportEnd,
		}
		body, err := json.Marshal(reqParams)
		require.NoError(t, err, "should be able to json encode request parameters")
		collectEndpoint := "/api/v1/datasources/prometheus/collect"
		t.Logf("Querying %s, currentTime: %s", collectEndpoint, currentTime)
		req := f.NewMeteringSVCPOSTRequest(collectEndpoint, body)
		result := req.Do()
		resp, err := result.Raw()
		t.Logf("Finishing querying %s, took: %s to finish", collectEndpoint, time.Now().UTC().Sub(currentTime))
		require.NoErrorf(t, err, "expected no errors triggering data collection, body: %v", string(resp))
	})
	return f.reportStart, f.reportEnd
}

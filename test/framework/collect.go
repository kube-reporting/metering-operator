package framework

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"github.com/operator-framework/operator-metering/pkg/operator"
)

// collectionSize is how much data is going to be scraped from Prometheus and
// imported
const collectionSize = 30 * time.Minute

func (f *Framework) CollectMetricsOnce(t *testing.T) (time.Time, time.Time) {
	t.Helper()
	f.collectOnce.Do(func() {
		// Use UTC, Prometheus uses UTCf for timestamps
		currentTime := time.Now().UTC()
		fiveMinutesAgo := currentTime.Add(-5 * time.Minute)

		// store the start/end so that future calls can immediately return the
		// same reportStart/reportEnd.

		// reportEnd is 5 minutes, to ensure data has scraped by Prometheus
		f.reportEnd = fiveMinutesAgo
		// reportStart is set to be before reportEnd by the size of the
		// collection we want to make.
		f.reportStart = f.reportEnd.Add(-collectionSize)

		reqParams := operator.CollectPromsumDataRequest{
			StartTime: f.reportStart,
			EndTime:   f.reportEnd,
		}
		body, err := json.Marshal(reqParams)
		require.NoError(t, err, "should be able to json encode request parameters")
		collectEndpoint := "/api/v1/datasources/prometheus/collect"
		t.Logf("currentTime: %s", currentTime)
		t.Logf("querying %s, with startTime: %s endTime: %s", collectEndpoint, f.reportStart, f.reportEnd)
		req := f.NewReportingOperatorSVCPOSTRequest(collectEndpoint, body)
		result := req.Do()
		resp, err := result.Raw()
		t.Logf("finished querying %s, took: %s to finish", collectEndpoint, time.Now().UTC().Sub(currentTime))
		require.NoErrorf(t, err, "expected no errors triggering data collection, body: %v", string(resp))
		var collectResp operator.CollectPromsumDataResponse
		err = json.Unmarshal(resp, &collectResp)
		require.NoError(t, err, "expected to unmarshal CollectPrometheusData response as JSON")
		t.Logf("CollectPromsumDataResponse: %s", spew.Sdump(collectResp))
		require.NotEmpty(t, collectResp.Results, "expected multiple import results")
		for _, result := range collectResp.Results {
			require.NotZerof(t, result.MetricsImportedCount, "expected metric import count for ReportDataSource %s to not be zero", result.ReportDataSource)
		}
	})
	return f.reportStart, f.reportEnd
}

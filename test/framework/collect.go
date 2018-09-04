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
		hourAndHalfAgo := currentTime.Add(-90 * time.Minute)

		// store the start/end so that future calls can immediately return the
		// same reportStart/reportEnd.

		// reportEnd is 1.5 hours ago, to ensure data has scraped by Prometheus
		f.reportEnd = hourAndHalfAgo
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
		t.Logf("Querying %s, currentTime: %s", collectEndpoint, currentTime)
		req := f.NewReportingOperatorSVCPOSTRequest(collectEndpoint, body)
		result := req.Do()
		resp, err := result.Raw()
		t.Logf("Finishing querying %s, took: %s to finish", collectEndpoint, time.Now().UTC().Sub(currentTime))
		require.NoErrorf(t, err, "expected no errors triggering data collection, body: %v", string(resp))
		var collectResp operator.CollectPromsumDataResponse
		err = json.Unmarshal(resp, &collectResp)
		require.NoError(t, err, "expected to unmarshal CollectPrometheusData response as JSON")
		t.Logf("CollectPromsumDataResponse: %s", spew.Sdump(collectResp))
	})
	return f.reportStart, f.reportEnd
}

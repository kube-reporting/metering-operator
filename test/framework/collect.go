package framework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"github.com/operator-framework/operator-metering/pkg/operator"
)

// collectionSize is how much data is going to be scraped from Prometheus and
// imported
const collectionSize = time.Hour

func (f *Framework) CollectMetricsOnce(t *testing.T) (time.Time, time.Time, operator.CollectPromsumDataResponse) {
	t.Helper()
	f.collectOnce.Do(func() {
		// Use UTC, Prometheus uses UTC for timestamps
		now := time.Now().UTC()

		// truncate the current time to the nearest hour
		nearestHour := now.Truncate(time.Hour)

		// set endTime 1 hour into the past to ensure it's far enough into
		// the past to have it's entire period elapsed
		f.reportEnd = nearestHour.Add(-time.Hour)
		// reportStart is set to be before reportEnd by the size of the
		// collection we want to make.
		f.reportStart = f.reportEnd.Add(-collectionSize)

		reqParams := operator.CollectPromsumDataRequest{
			StartTime: f.reportStart,
			EndTime:   f.reportEnd,
		}
		body, err := json.Marshal(reqParams)
		require.NoError(t, err, "should be able to json encode request parameters")
		collectEndpoint := fmt.Sprintf("/api/v1/datasources/prometheus/collect/%s", f.Namespace)
		t.Logf("currentTime: %s", now)
		t.Logf("querying %s, with startTime: %s endTime: %s", collectEndpoint, f.reportStart, f.reportEnd)
		respBody, respCode, err := f.ReportingOperatorPOSTRequest(collectEndpoint, body)
		require.Equal(t, http.StatusOK, respCode, "http response status code should be ok")
		require.NoErrorf(t, err, "expected no errors triggering data collection")
		t.Logf("***REMOVED***nished querying %s, took: %s to ***REMOVED***nish", collectEndpoint, time.Now().UTC().Sub(now))
		require.NoError(t, err, "reading response body should succeed")
		var collectResp operator.CollectPromsumDataResponse
		err = json.Unmarshal(respBody, &collectResp)
		require.NoError(t, err, "expected to unmarshal CollectPrometheusData response as JSON")
		t.Logf("CollectPromsumDataResponse: %s", spew.Sdump(collectResp))
		require.NotEmpty(t, collectResp.Results, "expected multiple import results")
		f.collectPromsumDataResponse = collectResp
	})
	return f.reportStart, f.reportEnd, f.collectPromsumDataResponse
}

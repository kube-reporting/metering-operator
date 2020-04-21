package reportingframework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"github.com/kubernetes-reporting/metering-operator/pkg/operator"
)

// collectionSize is how much data is going to be scraped from Prometheus and
// imported
const collectionSize = time.Hour

func (rf *ReportingFramework) CollectMetricsOnce(t *testing.T) (time.Time, time.Time, operator.CollectPrometheusMetricsDataResponse) {
	t.Helper()
	rf.collectOnce.Do(func() {
		// Use UTC, Prometheus uses UTC for timestamps
		now := time.Now().UTC()

		// truncate the current time to the nearest hour
		nearestHour := now.Truncate(time.Hour)

		// set endTime 1 hour into the past to ensure it's far enough into
		// the past to have it's entire period elapsed
		rf.reportEnd = nearestHour.Add(-time.Hour)
		// reportStart is set to be before reportEnd by the size of the
		// collection we want to make.
		rf.reportStart = rf.reportEnd.Add(-collectionSize)

		reqParams := operator.CollectPrometheusMetricsDataRequest{
			StartTime: rf.reportStart,
			EndTime:   rf.reportEnd,
		}
		body, err := json.Marshal(reqParams)
		require.NoError(t, err, "should be able to json encode request parameters")
		collectEndpoint := fmt.Sprintf("/api/v1/datasources/prometheus/collect/%s", rf.Namespace)
		t.Logf("currentTime: %s", now)
		t.Logf("querying %s, with startTime: %s endTime: %s", collectEndpoint, rf.reportStart, rf.reportEnd)
		respBody, respCode, err := rf.ReportingOperatorPOSTRequest(collectEndpoint, body)
		require.Equal(t, http.StatusOK, respCode, "http response status code should be ok")
		require.NoErrorf(t, err, "expected no errors triggering data collection")
		t.Logf("finished querying %s, took: %s to finish", collectEndpoint, time.Now().UTC().Sub(now))
		require.NoError(t, err, "reading response body should succeed")
		var collectResp operator.CollectPrometheusMetricsDataResponse
		err = json.Unmarshal(respBody, &collectResp)
		require.NoError(t, err, "expected to unmarshal CollectPrometheusData response as JSON")
		t.Logf("CollectPrometheusMetricsDataResponse: %s", spew.Sdump(collectResp))
		require.NotEmpty(t, collectResp.Results, "expected multiple import results")
		rf.collectPrometheusMetricsDataResponse = collectResp
	})
	return rf.reportStart, rf.reportEnd, rf.collectPrometheusMetricsDataResponse
}

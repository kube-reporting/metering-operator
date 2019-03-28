package framework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/operator-framework/operator-metering/pkg/operator"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/stretchr/testify/require"
)

func (f *Framework) StoreDataSourceData(t *testing.T, dataSourceName string, metrics []*prestostore.PrometheusMetric) error {
	params := operator.StorePromsumDataRequest(metrics)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/api/v1/datasources/prometheus/store/%s/%s", f.Namespace, dataSourceName)
	_, respCode, err := f.ReportingOperatorPOSTRequest(url, body)
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s", err)
	}
	require.Equal(t, http.StatusOK, respCode, "http response status code should be ok")

	return nil
}

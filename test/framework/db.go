package framework

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/operator-framework/operator-metering/pkg/operator"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
)

func (f *Framework) StoreDataSourceData(dataSourceName string, metrics []*prestostore.PrometheusMetric) error {
	params := operator.StorePromsumDataRequest(metrics)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/api/v1/datasources/prometheus/store/%s/%s", f.Namespace, dataSourceName)
	respBody, respCode, err := f.ReportingOperatorPOSTRequest(url, body)
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s", err)
	}
	if respCode != http.StatusOK {
		return fmt.Errorf("got http status %d when storing Metrics for ReportDataSource %s, body: %s", respCode, dataSourceName, string(respBody))
	}

	return nil
}

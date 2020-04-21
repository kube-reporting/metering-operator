package reportingframework

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubernetes-reporting/metering-operator/pkg/operator"
	"github.com/kubernetes-reporting/metering-operator/pkg/operator/prestostore"
)

func (rf *ReportingFramework) StoreDataSourceData(dataSourceName string, metrics []*prestostore.PrometheusMetric) error {
	params := operator.StorePrometheusMetricsDataRequest(metrics)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/api/v1/datasources/prometheus/store/%s/%s", rf.Namespace, dataSourceName)
	respBody, respCode, err := rf.ReportingOperatorPOSTRequest(url, body)
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s", err)
	}
	if respCode != http.StatusOK {
		return fmt.Errorf("got http status %d when storing Metrics for ReportDataSource %s, body: %s", respCode, dataSourceName, string(respBody))
	}

	return nil
}

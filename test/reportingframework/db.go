package reportingframework

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kube-reporting/metering-operator/pkg/operator"
	"github.com/kube-reporting/metering-operator/pkg/operator/prestostore"
)

const prometheusDataSourceAPIEndpointPrefix = "/api/v1/datasources/prometheus/store"

// StoreDataSourceData is a reportingframework method responsible for making
// the metering push API call to inject the list of @metrics prometheus metrics
// into a particular ReportDatasource custom resource database table.
func (rf *ReportingFramework) StoreDataSourceData(dataSourceName string, metrics []*prestostore.PrometheusMetric) error {
	params := operator.StorePrometheusMetricsDataRequest(metrics)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// build up the URL the post api request is expecting.
	// ex: $HOST/api/v1/datasources/prometheus/store/tflannag/namespace-cpu-request
	url := fmt.Sprintf("%s/%s/%s", prometheusDataSourceAPIEndpointPrefix, rf.Namespace, dataSourceName)
	respBody, respCode, err := rf.ReportingOperatorPOSTRequest(url, body)
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s", err)
	}
	if respCode != http.StatusOK {
		return fmt.Errorf("got http status %d when storing Metrics for ReportDataSource %s, body: %s", respCode, dataSourceName, string(respBody))
	}

	return nil
}

package framework

import (
	"encoding/json"
	"fmt"

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
	req := f.NewReportingOperatorSVCPOSTRequest(url, body)

	resp, err := req.Do().Raw()
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s body: %s", err, string(resp))
	}

	return nil
}

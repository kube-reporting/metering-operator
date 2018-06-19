package framework

import (
	"encoding/json"
	"fmt"

	"github.com/operator-framework/operator-metering/pkg/chargeback"
	"github.com/operator-framework/operator-metering/pkg/promcollector"
)

func (f *Framework) StoreDataSourceData(dataSourceName string, records []*promcollector.Record) error {
	params := chargeback.StorePromsumDataRequest(records)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/api/v1/datasources/prometheus/store/%s", dataSourceName)
	req := f.NewChargebackSVCPOSTRequest(url, body)

	resp, err := req.Do().Raw()
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s body: %s", err, string(resp))
	}

	return nil
}

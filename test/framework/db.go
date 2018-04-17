package framework

import (
	"encoding/json"
	"fmt"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

func (f *Framework) StoreDataSourceData(dataSourceName string, records []*chargeback.PromsumRecord) error {
	params := chargeback.StorePromsumDataRequest(records)
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/api/v1/datasources/prometheus/store/%s", dataSourceName)
	req := f.NewChargebackSVCPOSTRequest("chargeback", url, body)

	resp, err := req.Do().Raw()
	if err != nil {
		return fmt.Errorf("error storing datasource data: %s body: %s", err, string(resp))
	}

	return nil
}

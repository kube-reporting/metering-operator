package framework

import (
	"fmt"

	chargebackv1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
)

func (f *Framework) CreateChargebackReport(ns string, report *chargebackv1alpha1.Report) error {
	_, err := f.ChargebackClient.Reports(ns).Create(report)
	if err != nil {
		return fmt.Errorf("creating report %s failed, err %v", report.Name, err)
	}
	return nil
}

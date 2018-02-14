package framework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	chargebackv1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
)

func (f *Framework) CreateChargebackReport(ns string, report *chargebackv1alpha1.Report) error {
	_, err := f.ChargebackClient.Reports(ns).Create(report)
	return err
}

func (f *Framework) CreateChargebackScheduledReport(ns string, report *chargebackv1alpha1.ScheduledReport) error {
	_, err := f.ChargebackClient.ScheduledReports(ns).Create(report)
	return err
}

func (f *Framework) GetChargebackScheduledReport(ns, name string) (*chargebackv1alpha1.ScheduledReport, error) {
	return f.ChargebackClient.ScheduledReports(ns).Get(name, metav1.GetOptions{})
}

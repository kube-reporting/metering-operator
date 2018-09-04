package framework

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

func (f *Framework) GetMeteringReportGenerationQuery(name string) (*meteringv1alpha1.ReportGenerationQuery, error) {
	return f.MeteringClient.ReportGenerationQueries(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) WaitForMeteringReportGenerationQuery(t *testing.T, name string, pollInterval, timeout time.Duration) (*meteringv1alpha1.ReportGenerationQuery, error) {
	var reportQuery *meteringv1alpha1.ReportGenerationQuery
	return reportQuery, wait.Poll(pollInterval, timeout, func() (bool, error) {
		var err error
		reportQuery, err = f.GetMeteringReportGenerationQuery(name)
		if err != nil {
			if errors.IsNotFound(err) {
				t.Logf("ReportGenerationQuery %s does not exist yet", name)
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
}

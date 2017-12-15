package integration

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	chargebackv1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wellKnownReport = &chargebackv1alpha1.Report{
	ObjectMeta: metav1.ObjectMeta{
		Name: "pod-memory-usage-aws-billing-ci",
	},
	Spec: chargebackv1alpha1.ReportSpec{
		ReportingStart: metav1.Time{
			Time: time.Date(2017, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		ReportingEnd: metav1.Time{
			Time: time.Date(2017, 11, 30, 23, 59, 59, 0, time.UTC),
		},
		GenerationQueryName: "pod-memory-usage-aws-billing",
		RunImmediately:      true,
	},
}

func TestExampleReportsProduceData(t *testing.T) {
	deserializer := scheme.Codecs.UniversalDeserializer()

	tests := []struct {
		// name is the name of the sub test but also the name of the report.
		name       string
		reportFile string
	}{
		{
			name:       "node-cpu-usage",
			reportFile: "../../manifests/custom-resources/reports/node-cpu-usage.yaml",
		},
		{
			name:       "node-memory-usage",
			reportFile: "../../manifests/custom-resources/reports/node-memory-usage.yaml",
		},
		{
			name:       "pod-cpu-usage-by-namespace",
			reportFile: "../../manifests/custom-resources/reports/pod-cpu-usage-by-namespace.yaml",
		},
		{
			name:       "pod-cpu-usage-by-node",
			reportFile: "../../manifests/custom-resources/reports/pod-cpu-usage-by-node.yaml",
		},
		{
			name:       "pod-memory-usage-by-namespace",
			reportFile: "../../manifests/custom-resources/reports/pod-memory-usage-by-namespace.yaml",
		},
		{
			name:       "pod-memory-usage-by-node",
			reportFile: "../../manifests/custom-resources/reports/pod-memory-usage-by-node.yaml",
		},
		{
			name:       "pod-memory-usage-by-node-with-usage-percent",
			reportFile: "../../manifests/custom-resources/reports/pod-memory-usage-with-usage-percent.yaml",
		},
		// TODO(chancez): Add AWS Reports
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(test.reportFile)
			require.NoError(t, err, "expected manifest to exist")

			obj, _, err := deserializer.Decode(data, nil, nil)
			require.NoError(t, err, "expected to decode object")

			report := obj.(*chargebackv1alpha1.Report)
			require.NotNil(t, report, "report should not be nil")

			t.Logf("creating report %s", report.Name)
			err = testFramework.CreateChargebackReport(testFramework.Namespace, report)
			require.NoError(t, err, "creating report should succeed")

			defer func() {
				t.Logf("deleting report %s", report.Name)
				err := testFramework.ChargebackClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete report to succeed")
			}()

			query := map[string]string{
				"name":   test.name,
				"format": "json",
			}

			err = wait.Poll(time.Second*5, time.Minute*3, func() (bool, error) {
				req := testFramework.NewChargebackSVCRequest(testFramework.Namespace, "chargeback", "/api/v1/reports/get", query)
				result := req.Do()
				resp, err := result.Raw()
				if err != nil {
					return false, err
				}

				statusCode := new(int)
				result.StatusCode(statusCode)

				if *statusCode == http.StatusAccepted {
					t.Logf("report is still running")
					return false, nil
				}

				require.Equal(t, http.StatusOK, *statusCode, "http response status code should be ok")

				var reportResults []map[string]interface{}
				err = json.Unmarshal(resp, &reportResults)
				require.NoError(t, err, "expected to unmarshal response")
				assert.NotEqual(t, 0, len(reportResults), "reports should not return 0 results")
				return true, nil
			})
			assert.NoError(t, err, "expected getting report result to not timeout")
		})
	}
}

package deployframework

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/deploy"
)

func TestNewDeployerCon***REMOVED***g(t *testing.T) {
	df := &DeployFramework{}
	spec := metering.MeteringCon***REMOVED***gSpec{}
	testNamespace := "test-ns"
	testMeteringOpRepo := "test-repo-1"
	testMeteringOpTag := "test-tag-1"
	testReportingOpRepo := "test-repo-2"
	testReportingOpTag := "test-tag-2"
	cfg, err := df.NewDeployerCon***REMOVED***g(testNamespace, testMeteringOpRepo, testMeteringOpTag, testReportingOpRepo, testReportingOpTag, spec)
	require.NoError(t, err)

	expectedCfg := &deploy.Con***REMOVED***g{
		Namespace:       testNamespace,
		Repo:            testMeteringOpRepo,
		Tag:             testMeteringOpTag,
		Platform:        defaultPlatform,
		DeleteNamespace: defaultDeleteNamespace,
		ExtraNamespaceLabels: map[string]string{
			"name": testNamespaceLabel,
		},
		OperatorResources: nil,
		MeteringCon***REMOVED***g: &metering.MeteringCon***REMOVED***g{
			ObjectMeta: meta.ObjectMeta{
				Name:      meteringcon***REMOVED***gMetadataName,
				Namespace: testNamespace,
			},
			Spec: metering.MeteringCon***REMOVED***gSpec{
				ReportingOperator: &metering.ReportingOperator{
					Spec: &metering.ReportingOperatorSpec{
						Image: &metering.ImageCon***REMOVED***g{
							Repository: testReportingOpRepo,
							Tag:        testReportingOpTag,
						},
					},
				},
			},
		},
	}

	assert.Equalf(t, cfg, expectedCfg, "meteringcon***REMOVED***g should have reporting-operator image tag and namespace overridden")
}

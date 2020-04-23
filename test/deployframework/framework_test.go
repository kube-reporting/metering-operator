package deployframework

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/deploy"
)

func TestNewDeployerConfig(t *testing.T) {
	df := &DeployFramework{}
	spec := metering.MeteringConfigSpec{}
	testNamespace := "test-ns"
	testMeteringOpRepo := "test-repo-1"
	testMeteringOpTag := "test-tag-1"
	testReportingOpRepo := "test-repo-2"
	testReportingOpTag := "test-tag-2"
	cfg, err := df.NewDeployerConfig(testNamespace, testMeteringOpRepo, testMeteringOpTag, testReportingOpRepo, testReportingOpTag, spec)
	require.NoError(t, err)

	expectedCfg := &deploy.Config{
		Namespace:       testNamespace,
		Repo:            testMeteringOpRepo,
		Tag:             testMeteringOpTag,
		Platform:        defaultPlatform,
		DeleteNamespace: defaultDeleteNamespace,
		ExtraNamespaceLabels: map[string]string{
			"name": testNamespaceLabel,
		},
		OperatorResources: nil,
		MeteringConfig: &metering.MeteringConfig{
			ObjectMeta: meta.ObjectMeta{
				Name:      meteringconfigMetadataName,
				Namespace: testNamespace,
			},
			Spec: metering.MeteringConfigSpec{
				ReportingOperator: &metering.ReportingOperator{
					Spec: &metering.ReportingOperatorSpec{
						Image: &metering.ImageConfig{
							Repository: testReportingOpRepo,
							Tag:        testReportingOpTag,
						},
					},
				},
			},
		},
	}

	assert.Equalf(t, cfg, expectedCfg, "meteringconfig should have reporting-operator image tag and namespace overridden")
}

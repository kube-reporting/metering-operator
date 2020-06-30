package e2e

import (
	"github.com/kube-reporting/metering-operator/test/deployframework"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

func testManualMeteringInstall(
	t *testing.T,
	deployerCtx *deployframework.DeployerCtx,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
	catalogSourceName,
	catalogSourceNamespace,
	subscriptionChannel,
	testOutputPath string,
	expectInstallErrMsg []string,
	expectInstallErr bool,
	testInstallFunction InstallTestCase,
) {
	t.Parallel()
	deployerCtx.Logger.Infof("DeployerCtx: %+v", deployerCtx)

	rf, err := deployerCtx.Setup(deployerCtx.Deployer.InstallOLM, expectInstallErr)

	canSafelyRunTest := testhelpers.AssertCanSafelyRunReportingTests(t, err, expectInstallErr, expectInstallErrMsg)
	if canSafelyRunTest {
		t.Run(testInstallFunction.Name, func(t *testing.T) {
			testInstallFunction.TestFunc(t, rf)
		})

		deployerCtx.Logger.Infof("The %s test has finished running", testInstallFunction.Name)
	}

	err = deployerCtx.Teardown(deployerCtx.Deployer.UninstallOLM)
	assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}

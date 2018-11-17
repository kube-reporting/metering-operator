def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/deploy-continuous-upgrade.sh"
    testType   = "continuous-upgrade"
    kubeconfigCredentialsID = 'openshift-metering-ci-kubeconfig'
    deployPlatform = "openshift"
    alwaysSkipNamespaceCleanup = true
    uninstallMeteringBeforeInstall = false
}

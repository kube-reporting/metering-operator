def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/deploy-continuous-upgrade.sh"
    testType   = "continuous-upgrade"
    kubecon***REMOVED***gCredentialsID = 'openshift-metering-ci-kubecon***REMOVED***g'
    deployPlatform = "openshift"
    alwaysSkipNamespaceCleanup = true
    uninstallMeteringBeforeInstall = false
}

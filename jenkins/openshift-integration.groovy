def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/integration.sh"
    testType   = "integration"
    kubecon***REMOVED***gCredentialsID = 'openshift-metering-ci-kubecon***REMOVED***g'
    deployPlatform = "openshift"
    meteringHttpsAPI = true
}

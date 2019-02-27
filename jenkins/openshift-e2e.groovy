def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/e2e.sh"
    testType   = "e2e"
    kubecon***REMOVED***gCredentialsID = 'openshift-metering-ci-kubecon***REMOVED***g'
    deployPlatform = "openshift"
    meteringHttpsAPI = true
}

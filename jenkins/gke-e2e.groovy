def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/e2e-ci.sh"
    testType   = "e2e"
    kubecon***REMOVED***gCredentialsID = 'gke-metering-ci-kubecon***REMOVED***g'
    deployPlatform = "generic"
}

def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/e2e.sh"
    testType   = "e2e"
    kubeconfigCredentialsID = 'openshift-metering-ci-kubeconfig'
    deployPlatform = "openshift"
    meteringHttpsAPI = true
}

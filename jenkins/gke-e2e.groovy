def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/e2e-ci.sh"
    testType   = "e2e"
    kubeconfigCredentialsID = 'gke-metering-ci-kubeconfig'
    deployPlatform = "generic"
}

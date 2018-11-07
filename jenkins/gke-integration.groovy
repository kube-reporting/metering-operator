def testRunner = evaluate(readTrusted('jenkins/vars/testRunner.groovy'))

testRunner {
    testScript = "hack/integration-ci.sh"
    testType   = "integration"
    kubeconfigCredentialsID = 'gke-metering-ci-kubeconfig'
    deployPlatform = "generic"
}

def testRunner = evaluate(readTrusted('jenkins/testRunner.groovy'))

testRunner {
    testScript = "hack/integration-ci.sh"
    testType   = "integration"
}

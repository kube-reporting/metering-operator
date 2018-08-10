def testRunner = evaluate(readTrusted('jenkins/testRunner.groovy'))

testRunner {
    testScript = "hack/e2e-ci.sh"
    testType   = "e2e"
}

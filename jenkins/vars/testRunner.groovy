def call(body) {
    // evaluate the body block, and collect con***REMOVED***guration into the object
    def pipelineParams= [:]
    body.resolveStrategy = Closure.DELEGATE_FIRST
    body.delegate = pipelineParams
    body()

    // The rest is the re-usable declarative pipeline
    pipeline {
        parameters {
            string(name: 'DEPLOY_TAG', defaultValue: env.BRANCH_NAME, description: 'The image tag for all images deployed to use. Includes the integration-tests image which is used as the Jenkins executor.')
            string(name: 'OVERRIDE_NAMESPACE', defaultValue: '', description: 'If set, sets the namespace to deploy to')
            booleanParam(name: 'GENERIC', defaultValue: false, description: 'If true, run the con***REMOVED***gured tests against a GKE cluster using the generic con***REMOVED***g.')
            booleanParam(name: 'OPENSHIFT', defaultValue: false, description: 'If true, run the con***REMOVED***gured tests against a Openshift cluster using the Openshift con***REMOVED***g.')
            booleanParam(name: 'TECTONIC', defaultValue: false, description: 'If true, run the con***REMOVED***gured tests against a Openshift cluster using the Openshift con***REMOVED***g.')
        }
        agent {
            kubernetes {
                label "operator-metering-${pipelineParams.testType}-${params.DEPLOY_TAG}"
                instanceCap 2
                idleMinutes 0
                defaultContainer 'jnlp'
                yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-k8s: operator-metering-${pipelineParams.testType}
spec:
  containers:
  - name: metering-test-runner
    image: quay.io/coreos/chargeback-integration-tests:${params.DEPLOY_TAG}
    imagePullPolicy: Always
    command:
    - 'cat'
    tty: true
    """
            }
        }

        options {
            timestamps()
            overrideIndexTriggers(false)
            disableConcurrentBuilds()
            skipDefaultCheckout()
            buildDiscarder(logRotator(
                artifactDaysToKeepStr: '14',
                artifactNumToKeepStr: '30',
                daysToKeepStr: '14',
                numToKeepStr: '30',
            ))
        }

        environment {
            GOPATH                      = "/go"
            METERING_SRC_DIR            = "/go/src/github.com/operator-framework/operator-metering"
            DEPLOY_TAG                  = "${params.DEPLOY_TAG}"
            OUTPUT_DEPLOY_LOG_STDOUT    = "true"
            OUTPUT_TEST_LOG_STDOUT      = "true"
            OUTPUT_DIR                  = "test_output"
            METERING_CREATE_PULL_SECRET = "true"
            // use the OVERRIDE_NAMESPACE if speci***REMOVED***ed, otherwise set namespace to pre***REMOVED***x + DEPLOY_TAG
            METERING_NAMESPACE          = "${params.OVERRIDE_NAMESPACE ?: "metering-ci2-${pipelineParams.testType}-${params.DEPLOY_TAG}"}"
            SCRIPT                      = "${pipelineParams.testScript}"
            TEST_LOG_FILE               = "${pipelineParams.testType}-tests.log"
            TEST_TAP_FILE               = "${pipelineParams.testType}-tests.tap"
            DEPLOY_LOG_FILE             = "${pipelineParams.testType}-deploy.log"
            DEPLOY_POD_LOGS_LOG_FILE    = "${pipelineParams.testType}-pod.log"
            FINAL_POD_LOGS_LOG_FILE     = "${pipelineParams.testType}-***REMOVED***nal-pod.log"
            DOCKER_CREDS                = credentials('quay-coreos-jenkins-push')
        }

        stages {
            stage('Run Tests') {
                parallel {

                    stage('generic') {
                        when {
                            expression { return params.GENERIC }
                        }
                        environment {
                            KUBECONFIG                          = credentials('gke-metering-ci-kubecon***REMOVED***g')
                            TEST_OUTPUT_DIR                     = "${env.OUTPUT_DIR}/generic/tests"
                            TEST_OUTPUT_PATH                    = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}"
                            TEST_RESULT_REPORT_OUTPUT_DIRECTORY = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}/reports"
                            DEPLOY_PLATFORM                     = "generic"
                        }
                        steps {
                            runTests()
                        }
                        post {
                            always {
                                echo 'Capturing test TAP output'
                                step([$class: "TapPublisher", testResults: "${TEST_OUTPUT_DIR}/${TEST_TAP_FILE}", failIfNoResults: false, planRequired: false])
                            }
                            cleanup {
                                cleanup()
                            }
                        }
                    }

                    stage('tectonic') {
                        when {
                            expression { return params.TECTONIC }
                        }
                        environment {
                            KUBECONFIG                          = credentials('chargeback-ci-kubecon***REMOVED***g')
                            TEST_OUTPUT_DIR                     = "${env.OUTPUT_DIR}/tectonic/tests"
                            TEST_OUTPUT_PATH                    = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}"
                            TEST_RESULT_REPORT_OUTPUT_DIRECTORY = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}/reports"
                            DEPLOY_PLATFORM                     = "tectonic"
                        }
                        steps {
                            runTests()
                        }
                        post {
                            always {
                                echo 'Capturing test TAP output'
                                step([$class: "TapPublisher", testResults: "${TEST_OUTPUT_DIR}/${TEST_TAP_FILE}", failIfNoResults: false, planRequired: false])
                            }
                            cleanup {
                                cleanup()
                            }
                        }
                    }


                    stage('openshift') {
                        when {
                            expression { return params.OPENSHIFT }
                        }
                        environment {
                            KUBECONFIG                          = credentials('openshift-chargeback-ci-kubecon***REMOVED***g')
                            TEST_OUTPUT_DIR                     = "${env.OUTPUT_DIR}/openshift/tests"
                            TEST_OUTPUT_PATH                    = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}"
                            TEST_RESULT_REPORT_OUTPUT_DIRECTORY = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}/reports"
                            DEPLOY_PLATFORM                     = "openshift"
                        }
                        steps {
                            runTests()
                        }
                        post {
                            always {
                                echo 'Capturing test TAP output'
                                step([$class: "TapPublisher", testResults: "${TEST_OUTPUT_DIR}/${TEST_TAP_FILE}", failIfNoResults: false, planRequired: false])
                            }
                            cleanup {
                                cleanup()
                            }
                        }
                    }
                }
            }
        }
        post {
            always {
                container('jnlp') {
                    archiveArtifacts artifacts: "${env.OUTPUT_DIR}/**", onlyIfSuccessful: false, allowEmptyArchive: true
                }
            }
        }
    }
}

private def runTests() {
    echo "Running metering e2e tests"
    container('metering-test-runner') {
        ansiColor('xterm') {
            timeout(15) {
                sh '''#!/bin/bash -ex
                cd $METERING_SRC_DIR
                mkdir -p $TEST_OUTPUT_PATH $TEST_RESULT_REPORT_OUTPUT_DIRECTORY
                $SCRIPT
                '''
            }
        }
    }
}

private def cleanup() {
    container('metering-test-runner') {
        echo "Deleting namespace ${env.METERING_NAMESPACE}"
        sh 'set -e; cd $METERING_SRC_DIR && ./hack/delete-ns.sh'
    }
}

return this

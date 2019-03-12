def call(body) {
    // evaluate the body block, and collect con***REMOVED***guration into the object
    def pipelineParams= [:]
    body.resolveStrategy = Closure.DELEGATE_FIRST
    body.delegate = pipelineParams
    body()

    def commonPre***REMOVED***x = "${pipelineParams.deployPlatform}-${pipelineParams.testType}"
    def branchPre***REMOVED***x = "${commonPre***REMOVED***x}-${params.DEPLOY_TAG ?: env.BRANCH_NAME.toLowerCase()}"
    def defaultNamespace = "metering-ci2-${branchPre***REMOVED***x}"
    def prStatusContext = "jenkins/${commonPre***REMOVED***x}"

    def alwaysSkipNamespaceCleanup = pipelineParams.alwaysSkipNamespaceCleanup ?: false
    echo "alwaysSkipNamespaceCleanup: ${alwaysSkipNamespaceCleanup}"
    echo "common: ${commonPre***REMOVED***x}"
    echo "branchPre***REMOVED***x: ${branchPre***REMOVED***x}"

    echo 'Setting Github commit status'
    githubNotify context: prStatusContext, status: 'PENDING', description: "${pipelineParams.testType} tests started"

    // The rest is the re-usable declarative pipeline
    pipeline {
        parameters {
            string(name: 'DEPLOY_TAG', defaultValue: '', description: 'The image tag for all images deployed to use. Includes the integration-tests image which is used as the Jenkins executor. If unset, uses env.BRANCH_NAME')
            string(name: 'OVERRIDE_NAMESPACE', defaultValue: '', description: 'If set, sets the namespace to deploy to')
            booleanParam(name: 'SKIP_NS_CLEANUP', defaultValue: false, description: 'If true, skip cleaning up the namespace after running tests.')
        }
        agent {
            kubernetes {
                cloud 'gke-metering'
                label "gke-operator-metering-${branchPre***REMOVED***x}"
                instanceCap 2
                idleMinutes 0
                defaultContainer 'jnlp'
                yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-k8s: metering-${commonPre***REMOVED***x}
spec:
  containers:
  - name: metering-test-runner
    image: quay.io/coreos/metering-src:${params.DEPLOY_TAG ?: env.BRANCH_NAME}
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
            METERING_OPERATOR_DEPLOY_TAG = "${params.DEPLOY_TAG ?: env.BRANCH_NAME}"
            REPORTING_OPERATOR_DEPLOY_TAG = "${params.DEPLOY_TAG ?: env.BRANCH_NAME}"
            OUTPUT_DEPLOY_LOG_STDOUT    = "true"
            OUTPUT_TEST_LOG_STDOUT      = "true"
            OUTPUT_DIR                  = "test_output"
            METERING_CREATE_PULL_SECRET = "true"
            // use the OVERRIDE_NAMESPACE if speci***REMOVED***ed, otherwise set namespace to pre***REMOVED***x + BRANCH_NAME
            METERING_NAMESPACE          = "${params.OVERRIDE_NAMESPACE ?: defaultNamespace}"
            METERING_E2E_NAMESPACE      = "${METERING_NAMESPACE}"
            METERING_INTEGRATION_NAMESPACE = "${METERING_NAMESPACE}"
            SCRIPT                      = "${pipelineParams.testScript}"
            // we set CLEANUP_METERING_NAMESPACE to false and instead handle cleanup on
            // our own
            CLEANUP_METERING_NAMESPACE            = "true"
            DOCKER_CREDS                = credentials('quay-coreos-jenkins-push')
        }

        stages {
            stage('Run Tests') {
                environment {
                    KUBECONFIG                        = credentials("${pipelineParams.kubecon***REMOVED***gCredentialsID}")
                    TEST_OUTPUT_DIR                   = "${env.OUTPUT_DIR}/${commonPre***REMOVED***x}-tests"
                    TEST_OUTPUT_PATH                  = "${env.WORKSPACE}/${env.TEST_OUTPUT_DIR}"
                    DEPLOY_PLATFORM                   = "${pipelineParams.deployPlatform}"
                    METERING_HTTPS_API                = "${pipelineParams.meteringHttpsAPI ?: false}"
                    METERING_CREATE_PULL_SECRET       = "true"
                    UNINSTALL_METERING_BEFORE_INSTALL = "${(pipelineParams.uninstallMeteringBeforeInstall != null) ? pipelineParams.uninstallMeteringBeforeInstall : true}"
                }
                steps {
                    runScript()
                }
                post {
                    always {
                        echo 'Capturing test TAP output'
                        step([$class: "TapPublisher", testResults: "${TEST_OUTPUT_DIR}/**/*.tap", failIfNoResults: false, planRequired: false])
                    }
                    cleanup {
                        script {
                            if (alwaysSkipNamespaceCleanup || params.SKIP_NS_CLEANUP) {
                                echo 'Skipping namespace cleanup'
                            } ***REMOVED*** {
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
                script {
                    echo 'Updating Github PR status'
                    def status
                    def description
                    if (currentBuild.currentResult ==  "SUCCESS") {
                        status = "SUCCESS"
                        description = "All stages succeeded"
                    } ***REMOVED*** {
                        status = "FAILURE"
                        description = "Some stages failed"
                    }
                    githubNotify context: prStatusContext, status: status, description: description
                }
            }
        }
    }
}

private def runScript() {
    container('metering-test-runner') {
        ansiColor('xterm') {
            timeout(25) {
                sh '''#!/bin/bash -ex
                export METERING_CREATE_PULL_SECRET=true
                export DOCKER_USERNAME="${DOCKER_CREDS_USR:-}"
                export DOCKER_PASSWORD="${DOCKER_CREDS_PSW:-}"
                cd $METERING_SRC_DIR
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

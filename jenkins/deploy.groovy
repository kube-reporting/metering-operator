pipeline {
    parameters {
        string(name: 'DEPLOY_TAG', defaultValue: '', description: 'The image tag for all images deployed to use. Includes the integration-tests image which is used as the Jenkins executor. If unset, uses env.BRANCH_NAME')
        string(name: 'OVERRIDE_NAMESPACE', defaultValue: '', description: 'If set, sets the namespace to deploy to')
        booleanParam(name: 'GENERIC', defaultValue: false, description: '')
        booleanParam(name: 'OPENSHIFT', defaultValue: false, description: '')
    }
    agent {
        kubernetes {
            label "operator-metering-deploy-${params.DEPLOY_TAG ?: env.BRANCH_NAME}"
            instanceCap 2
            idleMinutes 0
            defaultContainer 'jnlp'
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-k8s: operator-metering-deploy
spec:
  containers:
  - name: metering-test-runner
    image: quay.io/coreos/metering-e2e:${params.DEPLOY_TAG ?: env.BRANCH_NAME}
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
        METERING_SRC_DIR            = "/go/src/github.com/operator-framework/operator-metering"
        DEPLOY_TAG                  = "${params.DEPLOY_TAG ?: env.BRANCH_NAME}"
        DISABLE_PROMSUM             = "false"
        DELETE_PVCS                 = "false"
        METERING_CREATE_PULL_SECRET = "true"
        // use the OVERRIDE_NAMESPACE if speci***REMOVED***ed, otherwise set namespace to pre***REMOVED***x + DEPLOY_TAG
        METERING_NAMESPACE          = "${params.OVERRIDE_NAMESPACE ?: "metering-ci2-deploy-${env.DEPLOY_TAG}"}"
        OUTPUT_DIR                  = "test_output"
        OUTPUT_PATH                 = "${env.WORKSPACE}/${env.OUTPUT_DIR}"
        DOCKER_CREDS                = credentials('quay-coreos-jenkins-push')
    }

    stages {
        stage('Deploy') {
            parallel {
                // Generic/GKE
                stage('generic') {
                    when {
                        expression { return params.GENERIC }
                    }
                    environment {
                        KUBECONFIG      = credentials('gke-metering-ci-kubecon***REMOVED***g')
                        DEPLOY_PLATFORM = "generic"
                    }
                    steps {
                        deploy()
                    }
                    post {
                        always {
                            captureLogs()
                        }
                    }
                }
                stage('openshift') {
                    when {
                        expression { return params.OPENSHIFT }
                    }
                    environment {
                        KUBECONFIG      = credentials('openshift-metering-ci-kubecon***REMOVED***g')
                        DEPLOY_PLATFORM = "openshift"
                    }
                    steps {
                        deploy()
                    }
                    post {
                        always {
                            captureLogs()
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

def deploy() {
    echo "Deploying metering"
    container('metering-test-runner') {
        ansiColor('xterm') {
            timeout(15) {
                sh '''#!/bin/bash -ex
                cd $METERING_SRC_DIR
                hack/deploy-continuous-upgrade.sh
                '''
            }
        }
    }
}

def captureLogs() {
    container('metering-test-runner') {
        sh '''#!/bin/bash -ex
        cd $METERING_SRC_DIR
        mkdir -p $OUTPUT_PATH/$DEPLOY_PLATFORM
        hack/capture-pod-logs.sh $METERING_NAMESPACE > $OUTPUT_PATH/$DEPLOY_PLATFORM/pod-logs.txt
        '''
    }
}

def isMasterBranch = env.BRANCH_NAME == "master"
def isPullRequest = env.BRANCH_NAME.startsWith("PR-")

def branchTag = env.BRANCH_NAME.toLowerCase()
def deployTag = "${branchTag}-${currentBuild.number}"

def dockerBuildArgs = '--no-cache'
if (isPullRequest) {
    dockerBuildArgs = ''
}

def podLabel = "operator-metering-build-${isPullRequest ? 'pr' : 'master'}"
def maxInstances = isPullRequest ? 5 : 2
def idleMin = isPullRequest ? 60: 15

pipeline {
    agent {
        kubernetes {
            label podLabel
            instanceCap maxInstances
            idleMinutes idleMin
            defaultContainer 'jnlp'
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    ${podLabel}: 'true'
spec:
  containers:
  - name: docker
    image: docker:stable-dind
    imagePullPolicy: Always
    command:
    - 'dockerd-entrypoint.sh'
    args:
    - '--storage-driver=overlay'
    tty: true
    securityContext:
      privileged: true
    volumeMounts:
    - name: var-lib-docker
      mountPath: /var/lib/docker
  volumes:
  - name: var-lib-docker
    emptyDir: {}
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

    parameters {
        booleanParam(name: 'REBUILD_HELM_OPERATOR', defaultValue: false, description: 'If true, rebuilds quay.io/coreos/helm-operator, otherwise pulls latest of the image.')
    }
    environment {
        GOPATH            = "${env.WORKSPACE}/go"
        METERING_SRC_DIR  = "${env.WORKSPACE}/go/src/github.com/operator-framework/operator-metering"
        USE_LATEST_TAG    = "${isMasterBranch}"
        USE_RELEASE_TAG   = "${isMasterBranch}"
        PUSH_RELEASE_TAG  = "${isMasterBranch}"
        BRANCH_TAG        = "${branchTag}"
        DEPLOY_TAG        = "${deployTag}"
        BRANCH_TAG_CACHE  = "${isMasterBranch}"
        DOCKER_BUILD_ARGS = "${dockerBuildArgs}"
    }
    stages {
        stage('Prepare') {
            environment {
                DOCKER_CREDS = credentials('quay-coreos-jenkins-push')
            }
            steps {
                container('docker') {
                    checkout([
                        $class: 'GitSCM',
                        branches: scm.branches,
                        extensions: scm.extensions + [[$class: 'RelativeTargetDirectory', relativeTargetDir: env.METERING_SRC_DIR]],
                        userRemoteCon***REMOVED***gs: scm.userRemoteCon***REMOVED***gs
                    ])

                    script {
                        // putting this in the environment block above wasn't working, so we use script and just assign to the env global
                        env.REBUILD_HELM_OPERATOR = "${params.REBUILD_HELM_OPERATOR || (!isPullRequest)}"
                    }

                    sh '''
                    apk update
                    apk add git bash
                    '''

                    echo "Authenticating to docker registry"
                    sh 'docker login -u $DOCKER_CREDS_USR -p $DOCKER_CREDS_PSW quay.io'

                    echo "Installing build dependencies"
                    sh '''#!/bin/bash
                    set -ex
                    apk add make go libc-dev curl jq zip python py-pip
                    pip install pyyaml
                    export HELM_VERSION=2.8.0
                    curl \
                        --silent \
                        --show-error \
                        --location \
                        "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
                        | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm \
                        && chmod +x /usr/local/bin/helm
                    helm init --client-only --skip-refresh
                    helm repo remove stable || true
                    '''
                }
            }
        }

        stage('Test') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        sh '''#!/bin/bash
                        set -e
                        set -o pipefail
                        make ci-validate
                        make test
                        '''
                    }
                }
            }
        }

        stage('Build') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''#!/bin/bash -ex
                            make docker-build-all -j 2 \
                                BRANCH_TAG_CACHE=$BRANCH_TAG_CACHE \
                                REBUILD_HELM_OPERATOR=$REBUILD_HELM_OPERATOR \
                                USE_LATEST_TAG=$USE_LATEST_TAG \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG
                            '''
                        }
                    }
                }
            }
        }

        stage('Tag') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''#!/bin/bash -ex
                            make docker-tag-all -j 2
                            '''
                        }
                    }
                }
            }
        }

        stage('Push') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''#!/bin/bash -ex
                            make docker-push-all -j 2 \
                                USE_LATEST_TAG=$USE_LATEST_TAG \
                                PUSH_RELEASE_TAG=$PUSH_RELEASE_TAG \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG
                            '''
                        }
                    }
                }
            }
        }
    }
}

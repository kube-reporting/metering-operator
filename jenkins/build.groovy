def isMasterBranch = env.BRANCH_NAME == "master"
def isPullRequest = env.BRANCH_NAME.startsWith("PR-")

def branchTag = env.BRANCH_NAME.toLowerCase()
def deployTag = "${branchTag}-${currentBuild.number}"

def dockerBuildArgs = '--no-cache'
if (isPullRequest) {
    dockerBuildArgs = ''
}

def podLabel = "gke-operator-metering-build-${branchTag}"
def maxInstances = isPullRequest ? 5 : 2
def idleMin = isPullRequest ? 60: 15

def prStatusContext = 'jenkins/build'

if (isPullRequest) {
    echo 'Setting Github PR status'
    githubNotify context: prStatusContext, status: 'PENDING', description: 'Build started'
}

pipeline {
    agent {
        kubernetes {
            cloud 'gke-metering'
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
    - '--storage-driver=overlay2'
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
        booleanParam(name: 'USE_IMAGEBUILDER', defaultValue: false, description: 'If true, uses github.com/openshift/imagebuilder as the Docker client')
    }
    environment {
        GOPATH            = "${env.WORKSPACE}/go"
        METERING_SRC_DIR  = "${env.WORKSPACE}/go/src/github.com/operator-framework/operator-metering"
        USE_LATEST_TAG    = "${isMasterBranch}"
        USE_RELEASE_TAG   = "${isMasterBranch}"
        PUSH_RELEASE_TAG  = "${isMasterBranch}"
        BRANCH_TAG        = "${branchTag}"
        DEPLOY_TAG        = "${deployTag}"
        DOCKER_BUILD_ARGS = "${dockerBuildArgs}"
    }
    stages {
        stage('Prepare') {
            environment {
                DOCKER_CREDS = credentials('quay-coreos-metering_ci-push')
            }
            steps {
                container('docker') {
                    checkout([
                        $class: 'GitSCM',
                        branches: scm.branches,
                        extensions: scm.extensions + [[$class: 'RelativeTargetDirectory', relativeTargetDir: env.METERING_SRC_DIR]],
                        userRemoteConfigs: scm.userRemoteConfigs
                    ])

                    script {
                        // putting this in the environment block above wasn't working, so we use script and just assign to the env global
                        env.REBUILD_HELM_OPERATOR = "${params.REBUILD_HELM_OPERATOR || (!isPullRequest)}"
                        env.USE_IMAGEBUILDER = "${params.USE_IMAGEBUILDER}"
                    }

                    sh '''
                    apk update
                    apk add bash make git
                    '''

                    script {
                        if (params.USE_IMAGEBUILDER) {
                            sh '''
                            apk add go libc-dev
                            mkdir -p $GOPATH
                            rm -rf $GOPATH/src/github.com/openshift/imagebuilder /usr/local/bin/imagebuilder
                            git clone https://github.com/chancez/imagebuilder $GOPATH/src/github.com/openshift/imagebuilder
                            cd $GOPATH/src/github.com/openshift/imagebuilder
                            git checkout copy_dockerignore_into_image
                            go build -o /usr/local/bin/imagebuilder github.com/openshift/imagebuilder/cmd/imagebuilder
                            chmod +x /usr/local/bin/imagebuilder
                            '''
                        }
                    }

                    echo "Authenticating to docker registry"
                    sh 'docker login -u $DOCKER_CREDS_USR -p $DOCKER_CREDS_PSW quay.io'
                }

            }
        }

        stage('Build builder image') {
            when {
                expression {
                    return true
                }
            }
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''
                            make metering-builder-docker-build \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG \
                                CHECK_GO_FILES=false
                            '''
                        }
                    }
                }
            }
        }

        stage('Test') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh 'make metering-e2e-docker-build CHECK_GO_FILES=false'
                            sh 'make ci-validate-docker CHECK_GO_FILES=false'
                            sh 'make test-docker CHECK_GO_FILES=false'
                        }
                    }
                }
            }
            post {
                success {
                    githubNotify context: prStatusContext, status: 'PENDING', description: 'Test stage passed'
                }
            }
        }

        stage('Build') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''
                            make docker-build-all \
                                REBUILD_HELM_OPERATOR=$REBUILD_HELM_OPERATOR \
                                USE_LATEST_TAG=$USE_LATEST_TAG \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG \
                                CHECK_GO_FILES=false
                            '''
                        }
                    }
                }
            }
            post {
                success {
                    githubNotify context: prStatusContext, status: 'PENDING', description: 'Build stage passed'
                }
            }
        }

        stage('Tag') {
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh 'make docker-tag-all CHECK_GO_FILES=false'
                        }
                    }
                }
            }
            post {
                success {
                    githubNotify context: prStatusContext, status: 'PENDING', description: 'Tag stage passed'
                }
            }
        }

        stage('Push builder image') {
            when {
                expression {
                    return isMasterBranch
                }
            }
            steps {
                dir(env.METERING_SRC_DIR) {
                    container('docker') {
                        ansiColor('xterm') {
                            sh '''
                            make docker-push IMAGE_NAME=quay.io/coreos/metering-builder \
                                USE_LATEST_TAG=$USE_LATEST_TAG \
                                PUSH_RELEASE_TAG=$PUSH_RELEASE_TAG \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG \
                                CHECK_GO_FILES=false
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
                            sh '''
                            make docker-push-all \
                                USE_LATEST_TAG=$USE_LATEST_TAG \
                                PUSH_RELEASE_TAG=$PUSH_RELEASE_TAG \
                                BRANCH_TAG=$BRANCH_TAG \
                                DEPLOY_TAG=$DEPLOY_TAG \
                                CHECK_GO_FILES=false
                            '''
                        }
                    }
                }
            }
            post {
                success {
                    githubNotify context: prStatusContext, status: 'PENDING', description: 'Push stage passed'
                }
            }
        }
    }
    post {
        always {
            script {
                if (isPullRequest) {
                    echo 'Updating Github PR status'
                    def status
                    def description
                    if (currentBuild.currentResult ==  "SUCCESS") {
                        status = "SUCCESS"
                        description = "All stages succeeded"
                    } else {
                        status = "FAILURE"
                        description = "Some stages failed"
                    }
                    githubNotify context: prStatusContext, status: status, description: description
                    slackSend channel: '#team-metering-ci', tokenCredentialId: 'team-metering-ci-slack-token', color: 'good', message: "<${env.BUILD_URL}|*${env.JOB_NAME} (build ${env.BUILD_NUMBER})*>\n:quay: built & pushed images with tags: ${env.BRANCH_TAG}, ${env.DEPLOY_TAG}"
                }
            }
        }
    }
}

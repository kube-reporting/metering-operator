properties([
    buildDiscarder(logRotator(
        artifactDaysToKeepStr: '14',
        artifactNumToKeepStr: '30',
        daysToKeepStr: '14',
        numToKeepStr: '30',
    )),
    disableConcurrentBuilds(),
    pipelineTriggers([]),
    parameters([
        booleanParam(name: 'BUILD_RELEASE', defaultValue: false, description: ''),
        booleanParam(name: 'USE_BRANCH_AS_TAG', defaultValue: false, description: ''),
        booleanParam(name: 'RUN_E2E_TESTS', defaultValue: true, description: 'If true, run e2e tests.'),
        booleanParam(name: 'SHORT_TESTS', defaultValue: false, description: 'If true, run tests with -test.short=true for running a subset of tests'),
        booleanParam(name: 'SKIP_DOCKER_STAGES', defaultValue: false, description: 'If true, skips docker build, tag and push'),
        booleanParam(name: 'SKIP_NAMESPACE_CLEANUP', defaultValue: false, description: 'If true, skips deleting the Kubernetes namespace at the end of the job'),
    ])
])

def isPullRequest = env.BRANCH_NAME.startsWith("PR-")
def isMasterBranch = env.BRANCH_NAME == "master"

def branchTag = env.BRANCH_NAME.toLowerCase()
def deployTag = "${branchTag}-${currentBuild.number}"
def chargebackNamespacePrefix = "chargeback-ci-${branchTag}"
def chargebackE2ENamespace = "${chargebackNamespacePrefix}-e2e"

def instanceCap = isMasterBranch ? 1 : 5
def podLabel = "kube-chargeback-build-${isMasterBranch ? 'master' : 'pr'}"

def awsBillingBucket = "team-chargeback"
def awsBillingBucketPrefix = "cost-usage-report/team-chargeback-chancez/"

echo "Params:\n${params}"

podTemplate(
    cloud: 'kubernetes',
    containers: [
        containerTemplate(
            alwaysPullImage: false,
            envVars: [],
            command: 'dockerd-entrypoint.sh',
            args: '--storage-driver=overlay',
            image: 'docker:dind',
            name: 'docker',
            privileged: true,
            ttyEnabled: true,
        ),
    ],
    volumes: [
        emptyDirVolume(
            mountPath: '/var/lib/docker',
            memory: false,
        ),
    ],
    idleMinutes: 15,
    instanceCap: 5,
    label: podLabel,
    name: podLabel,
) {
    node (podLabel) {
    timestamps {
        def runE2ETests = isMasterBranch || params.RUN_E2E_TESTS || (isPullRequest && pullRequest.labels.contains("run-e2e-tests"))
        def shortTests = params.SHORT_TESTS || (isPullRequest && pullRequest.labels.contains("run-short-tests"))

        def gopath = "${env.WORKSPACE}/go"
        def kubeChargebackDir = "${gopath}/src/github.com/coreos-inc/kube-chargeback"
        def testOutputDir = "test_output"
        def testOutputDirAbsolutePath = "${env.WORKSPACE}/${testOutputDir}"

        def e2eTestLogFile = 'e2e-tests.log'
        def e2eDeployLogFile = 'e2e-tests-deploy.log'
        def e2eTestTapFile = 'e2e-tests.tap'

        def dockerBuildArgs = ''
        if (isMasterBranch) {
            dockerBuildArgs = '--no-cache'
        }

        def gitCommit
        def gitTag

        try {
            container('docker'){

                stage('checkout') {
                    sh '''
                    apk update
                    apk add git bash jq zip python py-pip
                    pip install pyyaml
                    '''

                    checkout([
                        $class: 'GitSCM',
                        branches: scm.branches,
                        extensions: scm.extensions + [[$class: 'RelativeTargetDirectory', relativeTargetDir: kubeChargebackDir]],
                        userRemoteConfigs: scm.userRemoteConfigs
                    ])

                    gitCommit = sh(returnStdout: true, script: "cd ${kubeChargebackDir} && git rev-parse HEAD").trim()
                    gitTag = sh(returnStdout: true, script: "cd ${kubeChargebackDir} && git describe --tags --exact-match HEAD 2>/dev/null || true").trim()
                    echo "Git Commit: ${gitCommit}"
                    if (gitTag) {
                        echo "This commit has a matching git Tag: ${gitTag}"
                    }

                    if (params.BUILD_RELEASE) {
                        if (params.USE_BRANCH_AS_TAG) {
                            gitTag = branchTag
                        } else if (!gitTag) {
                            error "Unable to detect git tag"
                        }
                        deployTag = gitTag
                    }
                }
            }

            withCredentials([
                [$class: 'AmazonWebServicesCredentialsBinding', credentialsId: 'kube-chargeback-s3', accessKeyVariable: 'AWS_ACCESS_KEY_ID', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'],
                usernamePassword(credentialsId: 'quay-coreos-jenkins-push', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME'),
            ]) {
                withEnv([
                    "JENKINS_WORKSPACE=${env.WORKSPACE}",
                    "GOPATH=${gopath}",
                    "USE_LATEST_TAG=${isMasterBranch}",
                    "BRANCH_TAG=${branchTag}",
                    "BRANCH_TAG_CACHE=${isMasterBranch}",
                    "DEPLOY_TAG=${deployTag}",
                    "GIT_TAG=${gitTag}",
                    "DOCKER_BUILD_ARGS=${dockerBuildArgs}",
                    "CHARGEBACK_E2E_NAMESPACE=${chargebackE2ENamespace}",
                    "CHARGEBACK_SHORT_TESTS=${shortTests}",
                    "ENABLE_AWS_BILLING=false",
                    "AWS_BILLING_BUCKET=${awsBillingBucket}",
                    "AWS_BILLING_BUCKET_PREFIX=${awsBillingBucketPrefix}",
                    "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}",
                    "AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}",
                    "CLEANUP_CHARGEBACK=${!params.SKIP_NAMESPACE_CLEANUP}",
                ]){
                    container('docker'){
                        echo "Authenticating to docker registry"
                        sh 'docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD quay.io'

                        stage('install dependencies') {
                            // Build & install thrift
                            sh '''#!/bin/bash
                            set -e
                            apk add make go libc-dev curl
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

                            export KUBERNETES_VERSION=1.8.3
                            curl \
                                --silent \
                                --show-error \
                                --location \
                                "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
                                -o /usr/local/bin/kubectl \
                                 && chmod +x /usr/local/bin/kubectl
                            '''
                        }

                        dir(kubeChargebackDir) {
                            stage('test') {
                                sh '''#!/bin/bash
                                set -e
                                set -o pipefail
                                make ci-validate
                                make test
                                '''
                            }

                            stage('build') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker build"
                                } else if (!params.BUILD_RELEASE) {
                                    ansiColor('xterm') {
                                        sh '''#!/bin/bash -ex
                                        make docker-build-all -j 2 \
                                            BRANCH_TAG_CACHE=${BRANCH_TAG_CACHE} \
                                            USE_LATEST_TAG=${USE_LATEST_TAG} \
                                            BRANCH_TAG=${BRANCH_TAG}
                                        '''
                                    }
                                } else {
                                    // Images should already have been built if
                                    // we're doing a release build. In the tag
                                    // stage we will pull and tag these images
                                    echo "Release build, skipping building of images."
                                }
                            }

                            stage('tag') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker tag"
                                } else if (!params.BUILD_RELEASE) {
                                    ansiColor('xterm') {
                                        sh '''#!/bin/bash -ex
                                        make docker-tag-all -j 2 \
                                            IMAGE_TAG=${DEPLOY_TAG}
                                        '''
                                    }
                                } else {
                                    ansiColor('xterm') {
                                        sh '''#!/bin/bash -ex
                                        make docker-tag-all \
                                            PULL_TAG_IMAGE_SOURCE=true \
                                            IMAGE_TAG=${GIT_TAG}
                                        '''
                                    }
                                }
                            }

                            stage('push') {
                                if (params.SKIP_DOCKER_STAGES) {
                                    echo "Skipping docker push"
                                } else if (!params.BUILD_RELEASE) {
                                    sh '''#!/bin/bash -ex
                                    make docker-push-all -j 2 \
                                        USE_LATEST_TAG=${USE_LATEST_TAG} \
                                        BRANCH_TAG=${BRANCH_TAG}
                                    # Unset BRANCH_TAG so we don't push the same
                                    # image twice
                                    unset BRANCH_TAG
                                    make docker-push-all -j 2 \
                                        IMAGE_TAG=${DEPLOY_TAG}
                                        BRANCH_TAG=
                                    '''
                                } else {
                                    sh '''#!/bin/bash -ex
                                    make docker-push-all -j 2 \
                                        USE_LATEST_TAG=false \
                                        IMAGE_TAG=${GIT_TAG}
                                    '''
                                }
                            }

                            stage('release') {
                                if (params.BUILD_RELEASE) {
                                    sh '''#!/bin/bash -ex
                                    make release RELEASE_VERSION=${BRANCH_TAG}
                                    '''
                                    archiveArtifacts artifacts: 'tectonic-chargeback-*.zip', fingerprint: true, onlyIfSuccessful: true
                                } else {
                                    echo "Skipping release step, not a release"
                                }
                            }
                        }
                    }

                    stage('e2e tests') {
                        withCredentials([
                            [$class: 'FileBinding', credentialsId: 'chargeback-ci-kubeconfig', variable: 'TECTONIC_KUBECONFIG'],
                            [$class: 'FileBinding', credentialsId: 'openshift-chargeback-ci-kubeconfig', variable: 'OPENSHIFT_KUBECONFIG'],
                        ]) {
                            parallel "tectonic-e2e": {
                                if (runE2ETests) {
                                    echo "Running chargeback e2e tests"
                                    def myTestDir = "${testOutputDir}/tectonic_e2e"
                                    def myTestDirAbs = "${testOutputDirAbsolutePath}/tectonic_e2e"
                                    e2eRunner(kubeChargebackDir, [
                                        "CHARGEBACK_NAMESPACE=${CHARGEBACK_E2E_NAMESPACE}",
                                        "INSTALL_METHOD=direct",
                                        "DEPLOY_SCRIPT=deploy-ci.sh",
                                        "KUBECONFIG=${TECTONIC_KUBECONFIG}",
                                        "TEST_OUTPUT_DIR=${myTestDirAbs}",
                                        "TEST_LOG_FILE=${e2eTestLogFile}",
                                        "DEPLOY_LOG_FILE=${e2eDeployLogFile}",
                                        "TEST_TAP_FILE=${e2eTestTapFile}",
                                        "ENTRYPOINT=hack/e2e.sh",
                                    ])
                                    step([$class: "TapPublisher", testResults: "${myTestDir}/${e2eTestTapFile}", failIfNoResults: false, planRequired: false])
                                } else {
                                    echo "Non-master branch, skipping chargeback e2e tests"
                                }
                            }, "openshift-e2e": {
                                if (runE2ETests) {
                                    echo "Running chargeback e2e tests"
                                    def myTestDir = "${testOutputDir}/openshift_e2e"
                                    def myTestDirAbs = "${testOutputDirAbsolutePath}/openshift_e2e"
                                    try {
                                        e2eRunner(kubeChargebackDir, [
                                            "CHARGEBACK_NAMESPACE=${CHARGEBACK_E2E_NAMESPACE}",
                                            "INSTALL_METHOD=openshift-direct",
                                            "DEPLOY_SCRIPT=deploy-openshift-ci.sh",
                                            "SKIP_COPY_PULL_SECRET=false",
                                            "KUBECONFIG=${OPENSHIFT_KUBECONFIG}",
                                            "TEST_OUTPUT_DIR=${myTestDirAbs}",
                                            "TEST_LOG_FILE=${e2eTestLogFile}",
                                            "DEPLOY_LOG_FILE=${e2eDeployLogFile}",
                                            "TEST_TAP_FILE=${e2eTestTapFile}",
                                            "ENTRYPOINT=hack/e2e.sh",
                                        ])
                                    } catch (e) {
                                        echo "Openshift e2e failed: ${e}"
                                    }
                                    step([$class: "TapPublisher", testResults: "${myTestDir}/${e2eTestTapFile}", failIfNoResults: false, planRequired: false, failedTestsMarkBuildAsFailure: false])
                                } else {
                                    echo "Non-master branch, skipping chargeback e2e tests"
                                }
                            }, failFast: false
                        }
                    }
                }
            }
        } catch (e) {
            // If there was an exception thrown, the build failed
            echo "Build failed"
            currentBuild.result = "FAILED"
            throw e
        } finally {
            archiveArtifacts artifacts: "${testOutputDir}/**", onlyIfSuccessful: false
            container('docker') {
                sh '((docker ps -aq | xargs docker kill) || true) > /dev/null 2>&1'
            }
            cleanWs notFailBuild: true
        }
    }
} // timestamps end
} // podTemplate end

def e2eRunner(kubeChargebackDir, envVars) {
    withEnv(envVars) {
        container('docker'){
            dir(kubeChargebackDir) {
                sh 'kubectl config current-context'
                sh 'kubectl config get-contexts'
                try {
                    ansiColor('xterm') {
                        timeout(15) {
                            sh '''#!/bin/bash -ex
                            mkdir -p ${TEST_OUTPUT_DIR}
                            touch ${TEST_OUTPUT_DIR}/${DEPLOY_LOG_FILE}
                            touch ${TEST_OUTPUT_DIR}/${TEST_LOG_FILE}
                            tail -f ${TEST_OUTPUT_DIR}/${DEPLOY_LOG_FILE} &
                            tail -f ${TEST_OUTPUT_DIR}/${TEST_LOG_FILE} &
                            docker run \
                            -i --rm \
                            --env-file <(env | grep -E 'INSTALL_METHOD|TEST|DEPLOY|KUBECONFIG|AWS|CHARGEBACK|PULL_SECRET') \
                            -v "${JENKINS_WORKSPACE}:${JENKINS_WORKSPACE}" \
                            -v "${KUBECONFIG}:${KUBECONFIG}" \
                            -v "${TEST_OUTPUT_DIR}:/out" \
                            quay.io/coreos/chargeback-integration-tests:${DEPLOY_TAG} \
                            ${ENTRYPOINT}
                            '''
                        }
                    }
                } finally {
                    if (!params.SKIP_NAMESPACE_CLEANUP) {
                        sh '''#!/bin/bash -e
                        ./hack/delete-ns.sh
                        '''
                    }
                }
            }
        }
    }
}

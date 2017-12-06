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
    ])
])

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
            // resourceRequestCpu: '1750m',
            // resourceRequestMemory: '1500Mi',
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
    idleMinutes: 5,
    instanceCap: 5,
    label: 'kube-chargeback-build',
    name: 'kube-chargeback-build',
) {
    node ('kube-chargeback-build') {
        def gitCommit
        def gitTag
        def isMasterBranch = env.BRANCH_NAME == "master"

        try {
            withEnv([
                "GOPATH=${env.WORKSPACE}/go",
                "USE_LATEST_TAG=${isMasterBranch}",
                "BRANCH_TAG=${env.BRANCH_NAME}"
            ]){
                container('docker'){

                    def kubeChargebackDir = "${env.WORKSPACE}/go/src/github.com/coreos-inc/kube-chargeback"
                    stage('checkout') {
                        sh """
                        apk update
                        apk add git bash jq
                        """

                        def branches;
                        if (params.RELEASE_TAG) {
                            branches = params.RELEASE_TAG
                        } else {
                            branches = scm.branches
                        }
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
                    }

                    withCredentials([
                        usernamePassword(credentialsId: 'quay-coreos-jenkins-push', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME'),
                    ]) {

                        echo "Authenticating to docker registry"
                        // Run separately so variables are interpolated by groovy, note the double quotes
                        sh "docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD quay.io"
                    }

                    stage('install dependencies') {
                        // Build & install thrift
                        sh '''#!/bin/bash
                        set -e
                        apk add make go libc-dev curl
                        export HELM_VERSION=2.6.2
                        curl \
                            --silent \
                            --show-error \
                            --location \
                            "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
                            | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm \
                            && chmod +x /usr/local/bin/helm
                        helm init --client-only --skip-refresh
                        helm repo remove stable
                        '''
                    }

                    dir(kubeChargebackDir) {
                        stage('test') {
                            sh """#!/bin/bash
                            make k8s-verify-codegen
                            """
                        }
                        if (params.BUILD_RELEASE) {
                            if (!gitTag) {
                                error "Unable to detect git tag"
                            }
                            stage('tag') {
                                ansiColor('xterm') {
                                    sh """#!/bin/bash
                                    make docker-tag-all \
                                        PULL_TAG_IMAGE_SOURCE=true \
                                        IMAGE_TAG=${gitTag}
                                    """
                                }
                            }
                            stage('push') {
                                sh """
                                make docker-push-all -j 2 \
                                    USE_LATEST_TAG=false \
                                    IMAGE_TAG=${gitTag}
                                """
                            }
                        } else {
                            stage('build') {
                                ansiColor('xterm') {
                                    sh """#!/bin/bash
                                    make docker-build-all -j 2 \
                                        USE_LATEST_TAG=${USE_LATEST_TAG} \
                                        BRANCH_TAG=${BRANCH_TAG}
                                    """
                                }
                            }

                            stage('push') {
                                sh """
                                make docker-push-all -j 2 \
                                    USE_LATEST_TAG=${USE_LATEST_TAG} \
                                    BRANCH_TAG=${BRANCH_TAG}
                                """
                            }
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
            cleanWs notFailBuild: true
            // notifyBuild(currentBuild.result)
        }
    }
}

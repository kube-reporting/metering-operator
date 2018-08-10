@Library('shared-libraries')
def aborter = new abortPreviousBuilds()
aborter.abortPreviousBuilds()

def isPullRequest = env.BRANCH_NAME.startsWith("PR-")
def isMasterBranch = env.BRANCH_NAME == "master"

def skipBuildLabel = (isPullRequest && pullRequest.labels.contains("skip-build"))
def skipE2ELabel = (isPullRequest && pullRequest.labels.contains("skip-e2e"))
def skipIntegrationLabel = (isPullRequest && pullRequest.labels.contains("skip-integration"))

def skipTectonic = (isPullRequest && pullRequest.labels.contains("skip-tectonic"))
def skipOpenshift = (isPullRequest && pullRequest.labels.contains("skip-openshift"))
def skipGke = (isPullRequest && pullRequest.labels.contains("skip-gke"))

pipeline {
    agent none
    parameters {
        booleanParam(name: 'BUILD', defaultValue: true, description: 'If true, builds and pushes the metering docker images')
        string(name: 'OVERRIDE_BRANCH_NAME', defaultValue: '', description: 'Branch to build. If unset, uses the current branch.')

        booleanParam(name: 'INTEGRATION', defaultValue: true, description: 'If true, then integration tests will run against the results of the build.')
        booleanParam(name: 'E2E', defaultValue: true, description: 'If true, then e2e tests will run against the results of the build.')

        booleanParam(name: 'GENERIC', defaultValue: true, description: 'If true, run the configured tests against a GKE cluster using the generic config.')
        booleanParam(name: 'OPENSHIFT', defaultValue: true, description: 'If true, run the configured tests against a Openshift cluster using the Openshift config.')
        booleanParam(name: 'TECTONIC', defaultValue: true, description: 'If true, run the configured tests against a Openshift cluster using the Openshift config.')
    }
    triggers {
        issueCommentTrigger('.*jenkins rebuild.*')
    }
    options {
        timestamps()
        skipDefaultCheckout()
        buildDiscarder(logRotator(
            artifactDaysToKeepStr: '14',
            artifactNumToKeepStr: '30',
            daysToKeepStr: '14',
            numToKeepStr: '30',
        ))
    }

    environment {
        // use the parameter branch name for the build and test jobs if set, otherwise default to the pipeline branch for this job
        TARGET_BRANCH = "${params.OVERRIDE_BRANCH_NAME ?: env.BRANCH_NAME}"
    }

    stages {
        stage('Setup') {
            steps {
                script {
                    if (isPullRequest) {
                        echo "Github PR labels: ${pullRequest.labels.join(',')}"
                    }

                }
            }
        }
        stage('Build') {
            when {
                expression {
                    return params.BUILD && !skipBuildLabel

                }
            }
            steps {
                echo "Building and pushing metering docker images"
                build job: "metering/operator-metering-build/${env.TARGET_BRANCH}"
            }
        }

        stage('Test') {
            parallel {
                stage("integration") {
                    when {
                        expression {
                            return params.INTEGRATION && !skipIntegrationLabel
                        }
                    }
                    steps {
                        echo "Running metering integration tests"
                        build job: "metering/operator-metering-integration/${env.TARGET_BRANCH}", parameters: [
                            string(name: 'DEPLOY_TAG', value: skipBuildLabel ? "master" : env.TARGET_BRANCH),
                            booleanParam(name: 'GENERIC', value: params.GENERIC && !skipGke),
                            booleanParam(name: 'OPENSHIFT', value: params.OPENSHIFT && !skipOpenshift),
                            booleanParam(name: 'TECTONIC', value: params.TECTONIC && !skipTectonic),
                        ]
                    }
                }
                stage("e2e") {
                    when {
                        expression {
                            return params.E2E && !skipE2ELabel
                        }
                    }
                    steps {
                        echo "Running metering e2e tests"
                        build job: "metering/operator-metering-e2e/${env.TARGET_BRANCH}", parameters: [
                            string(name: 'DEPLOY_TAG', value: skipBuildLabel ? "master" : env.TARGET_BRANCH),
                            booleanParam(name: 'GENERIC', value: params.GENERIC && !skipGke),
                            booleanParam(name: 'OPENSHIFT', value: params.OPENSHIFT && !skipOpenshift),
                            booleanParam(name: 'TECTONIC', value: params.TECTONIC && !skipTectonic),
                        ]
                    }
                }
            }
        }
    }
}

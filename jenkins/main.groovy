pipeline {
    agent none
    parameters {
        booleanParam(name: 'BUILD', defaultValue: true, description: 'If true, builds and pushes the metering docker images')
        string(name: 'OVERRIDE_BRANCH_NAME', defaultValue: '', description: 'Branch to build. If unset, uses the current branch.')

        booleanParam(name: 'INTEGRATION', defaultValue: true, description: 'If true, then integration tests will run against the results of the build.')
        booleanParam(name: 'E2E', defaultValue: true, description: 'If true, then e2e tests will run against the results of the build.')

        booleanParam(name: 'GENERIC', defaultValue: true, description: 'If true, run the con***REMOVED***gured tests against a GKE cluster using the generic con***REMOVED***g.')
        booleanParam(name: 'OPENSHIFT', defaultValue: true, description: 'If true, run the con***REMOVED***gured tests against a Openshift cluster using the Openshift con***REMOVED***g.')
        booleanParam(name: 'TECTONIC', defaultValue: true, description: 'If true, run the con***REMOVED***gured tests against a Openshift cluster using the Openshift con***REMOVED***g.')
    }
    options {
        timestamps()
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
        // use the parameter branch name for the build and test jobs if set, otherwise default to the pipeline branch for this job
        TARGET_BRANCH = "${params.OVERRIDE_BRANCH_NAME ?: env.BRANCH_NAME}"
    }

    stages {

        stage('Build') {
            when {
                equals expected: true, actual: params.BUILD
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
                        equals expected: true, actual: params.INTEGRATION
                    }
                    steps {
                        echo "Running metering integration tests"
                        build job: "metering/operator-metering-integration/${env.TARGET_BRANCH}", parameters: [
                            string(name: 'DEPLOY_TAG', value: env.TARGET_BRANCH),
                            booleanParam(name: 'GENERIC', value: params.GENERIC),
                            booleanParam(name: 'OPENSHIFT', value: params.OPENSHIFT),
                            booleanParam(name: 'TECTONIC', value: params.TECTONIC),
                        ]
                    }
                }
                stage("e2e") {
                    when {
                        equals expected: true, actual: params.E2E
                    }
                    steps {
                        echo "Running metering e2e tests"
                        build job: "metering/operator-metering-e2e/${env.TARGET_BRANCH}", parameters: [
                            string(name: 'DEPLOY_TAG', value: env.TARGET_BRANCH),
                            booleanParam(name: 'GENERIC', value: params.GENERIC),
                            booleanParam(name: 'OPENSHIFT', value: params.OPENSHIFT),
                            booleanParam(name: 'TECTONIC', value: params.TECTONIC),
                        ]
                    }
                }
            }
        }
    }
}

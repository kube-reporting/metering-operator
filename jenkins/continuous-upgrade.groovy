def notifyBuild = evaluate readTrusted('jenkins/vars/notifyBuild.groovy')

notifyBuild("STARTED")
pipeline {
    parameters {
        string(name: 'DEPLOY_TAG', defaultValue: '', description: 'The image tag for all images deployed to use. If unset, uses env.BRANCH_NAME')
        string(name: 'OVERRIDE_NAMESPACE', defaultValue: '', description: 'If set, sets the namespace to deploy to. If unset, the namespace is metering-ci2-continuous-upgrade-$env.BRANCH_NAME.')
        booleanParam(name: 'OPENSHIFT', defaultValue: true, description: '')
    }
    agent any
    triggers {
        cron('0 0 * * *') // Every night at midnight
        upstream(upstreamProjects: "metering/operator-metering-build/${env.BRANCH_NAME}", threshold: hudson.model.Result.SUCCESS)
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
        DEPLOY_TAG         = "${params.DEPLOY_TAG ?: env.BRANCH_NAME}"
        METERING_NAMESPACE = "${params.OVERRIDE_NAMESPACE ?: "metering-ci2-continuous-upgrade-${env.BRANCH_NAME}"}"
    }
    stages {
        stage('Deploy/Upgrade') {
            when {
                branch 'master'
            }
            steps {
                echo "Deploying/Upgrading namespace ${METERING_NAMESPACE}"
                build job: "metering/operator-metering-deploy/master", parameters: [
                    string(name: 'DEPLOY_TAG', value: env.DEPLOY_TAG),
                    string(name: 'OVERRIDE_NAMESPACE', value: env.METERING_NAMESPACE),
                    booleanParam(name: 'OPENSHIFT', value: params.OPENSHIFT),
                ]
            }
        }
    }
    post {
        always {
            script {
                notifyBuild(currentBuild.currentResult)
            }
        }
    }
}

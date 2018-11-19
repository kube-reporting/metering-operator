def notifyBuild = evaluate readTrusted('jenkins/vars/notifyBuild.groovy')

def prStatusContext = 'jenkins/continuous-upgrade'

githubNotify context: prStatusContext, status: 'PENDING', description: 'continuous-upgrade started'
notifyBuild("STARTED")
pipeline {
    parameters {
        string(name: 'DEPLOY_TAG', defaultValue: '', description: 'The image tag for all images deployed to use. If unset, uses env.BRANCH_NAME')
    }
    agent any
    triggers {
        cron('0 0 * * *') // Every night at midnight
        upstream(upstreamProjects: "metering/operator-metering-main/${env.BRANCH_NAME}", threshold: hudson.model.Result.SUCCESS)
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
    }
    stages {
        stage('Deploy/Upgrade') {
            when {
                branch 'master'
            }
            steps {
                build job: "metering/operator-metering-openshift-continuous-upgrade/${env.BRANCH_NAME}", parameters: [
                    string(name: 'DEPLOY_TAG', value: env.DEPLOY_TAG),
                ]
            }
            post {
                success {
                    githubNotify context: prStatusContext, status: 'SUCCESS', description: 'continuous-upgrade passed'
                }
            }
        }
    }
    post {
        always {
            script {
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
                notifyBuild(currentBuild.currentResult)
            }
        }
    }
}

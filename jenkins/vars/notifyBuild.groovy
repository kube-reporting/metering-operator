def call(String buildStatus = 'STARTED') {
    def buildColor = ''
    def buildEmoji = ''

    // build status of null means success
    buildStatus =  buildStatus ?: 'SUCCESS'

    // Override default values based on build status
    switch (buildStatus) {
        case 'STARTED':
            buildColor = 'warning'
            buildEmoji = ':open_mouth:'
            break;
        case 'SUCCESS':
            buildColor = 'good'
            buildEmoji = ':smirk:'
            break;
        default:
            buildColor = 'danger'
            buildEmoji = ':cold_sweat:'
            break;
    }


    slackSend(
        channel: '#team-metering-ci',
        color: buildColor,
        message: """
            |*<${env.BUILD_URL}|${env.JOB_NAME} build ${env.BUILD_NUMBER}>*
            |${buildEmoji} Build ${buildStatus}.
            |branch: ${env.BRANCH_NAME}
            """.stripMargin(),
        tokenCredentialId: 'team-metering-ci-slack-token'
    )
}

return this

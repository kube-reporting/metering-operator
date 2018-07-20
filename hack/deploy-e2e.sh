#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")

# Used in deploy.sh
export DOCKER_USERNAME="$DOCKER_CREDS_USR"
export DOCKER_PASSWORD="$DOCKER_CREDS_PSW"

export DISABLE_PROMSUM=true
export METERING_CREATE_PULL_SECRET=true
"$DIR/deploy-custom.sh"

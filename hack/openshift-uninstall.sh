#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

export DEPLOY_PLATFORM=openshift
"${DIR}/uninstall.sh"

#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/uninstall.sh"

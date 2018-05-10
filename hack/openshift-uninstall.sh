#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/uninstall.sh"

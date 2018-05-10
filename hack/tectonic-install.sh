#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..

export DEPLOY_PLATFORM=tectonic
"${ROOT_DIR}/hack/install.sh"

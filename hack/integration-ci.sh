#!/bin/bash
set -e

: "${METERING_NAMESPACE:?}"
: "${KUBECONFIG:?}"
: "${DEPLOY_TAG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-deploy-e2e.sh}"
"$ROOT_DIR/hack/integration.sh"

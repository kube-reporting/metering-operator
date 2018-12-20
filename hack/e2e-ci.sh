#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${METERING_NAMESPACE:?}"
: "${KUBECONFIG:?}"
: "${DEPLOY_TAG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-deploy-e2e.sh}"
"$ROOT_DIR/hack/e2e.sh"

#!/bin/bash
set -e

export DELETE_PVCS=${DELETE_PVCS:-true}

: "${CUSTOM_METERING_CR_FILE:?Must set \$CUSTOM_METERING_CR_FILE}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

echo "Deploying"
"${ROOT_DIR}/hack/deploy.sh"

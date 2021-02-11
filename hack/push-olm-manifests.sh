#!/bin/bash
set -e
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

IMAGE_METERING_ANSIBLE_OPERATOR_REGISTRY="${1:?}"
MANIFEST_BUNDLE="${2:-manifests/deploy/openshift/olm/bundle}"
CONTAINER_RUNTIME=${CONTAINER_RUNTIME:-podman}

echo "Building and pushing the operator manifest bundle to ${IMAGE_METERING_ANSIBLE_OPERATOR_REGISTRY}"
${CONTAINER_RUNTIME} build -f "${ROOT_DIR}"/olm_deploy/Dockerfile.registry -t "${IMAGE_METERING_ANSIBLE_OPERATOR_REGISTRY}" --build-arg MANIFEST_BUNDLE="${MANIFEST_BUNDLE}" .
${CONTAINER_RUNTIME} push "${IMAGE_METERING_ANSIBLE_OPERATOR_REGISTRY}"

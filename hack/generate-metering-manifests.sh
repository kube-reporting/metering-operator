#!/bin/bash

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

current_version=${1:-4.8}

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

msg "Generating Openshift Manifests"
"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$OCP_INSTALLER_MANIFESTS_DIR" \
    "$OCP_OLM_MANIFESTS_DIR" \
    "$ROOT_DIR/charts/metering-ansible-operator/values.yaml"

msg "Generating Upstream Manifests"
"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$UPSTREAM_INSTALLER_MANIFESTS_DIR" \
    "$UPSTREAM_OLM_MANIFESTS_DIR" \
    "$ROOT_DIR/charts/metering-ansible-operator/upstream-values.yaml"

msg "Generating Openshift Bundle"
${OPM_BIN} alpha bundle generate \
    --directory="${OCP_OLM_BUNDLE_DIR}/${current_version}" \
    --output-dir="${OCP_BUNDLE_DIR}" \
    --default="${current_version}" \
    --channels "${current_version}" \
    --package metering-ocp &&
    mv bundle.Dockerfile Dockerfile.bundle &&
    find "${OCP_BUNDLE_DIR}" -type f ! -name '*.yaml' -delete

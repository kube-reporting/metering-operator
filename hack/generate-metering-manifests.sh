#!/bin/bash

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

TMPDIR="$(mktemp -d)"
trap "rm -rf $TMPDIR" EXIT

echo "Generating Openshift Manifests"

"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$OCP_INSTALLER_MANIFESTS_DIR" \
    "$OCP_OLM_MANIFESTS_DIR"

echo "Generating Upstream Manifests"
"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$UPSTREAM_INSTALLER_MANIFESTS_DIR" \
    "$UPSTREAM_OLM_MANIFESTS_DIR" \
    "$ROOT_DIR/charts/metering-ansible-operator/upstream-values.yaml"

echo "Generating OCP Testing Manifests"

"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$OCP_TESTING_INSTALLER_MANIFESTS_DIR" \
    "$OCP_TESTING_OLM_MANIFESTS_DIR" \
    "$ROOT_DIR/charts/metering-ansible-operator/ocp-testing-values.yaml"

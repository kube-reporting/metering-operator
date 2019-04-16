#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are set, will be used as the paths to
# files containing override values for rendering the manifests to the output
# directory.

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:?}"
: "${CUSTOM_OLM_OVERRIDE_VALUES:?}"

echo "Using $INSTALLER_MANIFESTS_DIR as metering-operator manifest output dir and $OLM_MANIFESTS_DIR as OLM manifest output directory"
echo

# openshift
echo
echo "Creating Openshift deploy manifests"
"$ROOT_DIR/hack/create-deploy-manifests.sh" \
    "$INSTALLER_MANIFESTS_DIR" \
    "$DEPLOY_MANIFESTS_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-helm-operator-values.yaml" \
    "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES}"

echo
echo "Creating Openshift OLM manifests"
"$ROOT_DIR/hack/create-olm-manifests.sh" \
    "$INSTALLER_MANIFESTS_DIR" \
    "$OLM_MANIFESTS_DIR" \
    "$DEPLOY_MANIFESTS_DIR/common-olm-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-olm-values.yaml" \
    "${CUSTOM_OLM_OVERRIDE_VALUES}"

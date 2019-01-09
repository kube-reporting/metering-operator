#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are set, will be used as the paths to
# files containing override values for rendering the manifests to the output
# directory.

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

# By default, we output into the deploy directory, but this can be overridden
: "${MANIFEST_OUTPUT_DIR:=$DEPLOY_MANIFESTS_DIR}"
: "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:?}"
: "${CUSTOM_ALM_OVERRIDE_VALUES:?}"

echo "Using $MANIFEST_OUTPUT_DIR as output directory"
echo

# openshift
echo
echo "Creating Openshift deploy manifests"
"$ROOT_DIR/hack/create-deploy-manifests.sh" \
    "$MANIFEST_OUTPUT_DIR/openshift/helm-operator" \
    "$DEPLOY_MANIFESTS_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-helm-operator-values.yaml" \
    "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES}"

echo
echo "Creating Openshift ALM manifests"
"$ROOT_DIR/hack/create-alm-manifests.sh" \
    "$MANIFEST_OUTPUT_DIR/openshift/helm-operator" \
    "$MANIFEST_OUTPUT_DIR/openshift/alm" \
    "$DEPLOY_MANIFESTS_DIR/common-alm-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-alm-values.yaml" \
    "${CUSTOM_ALM_OVERRIDE_VALUES}"

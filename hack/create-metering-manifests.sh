#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are unspeci***REMOVED***ed, then this script
# requires 1 argument, the image tag for the helm-operator pod. Additionally,
# a second optional argument can be provided to override where the manifests
# should be output, by default it outputs everything into the
# deploy/manifests/openshift directory.
#
# If both of those environment variables are set, then this script takes no
# arguments and the environment variables will be used as the paths to ***REMOVED***les
# containing override values for rendering the manifests to the output
# directory. In this scenario, a ***REMOVED***rst optional argument can be provided to
# control the manifests output directory.

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

TMPDIR="$(mktemp -d)"

IMAGE_TAG=""

if [[ -z "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES-}" && -z "${CUSTOM_ALM_OVERRIDE_VALUES-}" ]]; then
    : "${1?"Usage: $0 IMAGE_TAG"}"
    echo "Using $1 as IMAGE_TAG"
    IMAGE_TAG="$1"
***REMOVED***

# By default, we output into the deploy directory, but this can be overridden
: "${MANIFEST_OUTPUT_DIR:=$DEPLOY_MANIFESTS_DIR}"

echo "Using $MANIFEST_OUTPUT_DIR as output directory"
echo

trap "rm -rf $TMPDIR" EXIT

HELM_OPERATOR_OVERRIDE_VALUES_FILE=""
ALM_OVERRIDE_VALUES_FILE=""

if [ -n "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES-}" ]; then
    HELM_OPERATOR_OVERRIDE_VALUES_FILE="$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
***REMOVED***
    HELM_OPERATOR_OVERRIDE_VALUES_FILE="$TMPDIR/override-helm-operator-values.yaml"
    "$ROOT_DIR/hack/render-helm-operator-override-values.sh" "$IMAGE_TAG" > "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"
***REMOVED***

if [ -n "${CUSTOM_ALM_OVERRIDE_VALUES-}" ]; then
    ALM_OVERRIDE_VALUES_FILE="$CUSTOM_ALM_OVERRIDE_VALUES"
***REMOVED***
    ALM_OVERRIDE_VALUES_FILE="$TMPDIR/override-alm-values.yaml"
    "$ROOT_DIR/hack/render-alm-override-values.sh" "$IMAGE_TAG" > "$ALM_OVERRIDE_VALUES_FILE"
***REMOVED***

# openshift
echo
echo "Creating Openshift deploy manifests"
"$ROOT_DIR/hack/create-deploy-manifests.sh" \
    "$MANIFEST_OUTPUT_DIR/openshift/helm-operator" \
    "$DEPLOY_MANIFESTS_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-helm-operator-values.yaml" \
    "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"

echo
echo "Creating Openshift ALM manifests"
"$ROOT_DIR/hack/create-alm-manifests.sh" \
    "$MANIFEST_OUTPUT_DIR/openshift/helm-operator" \
    "$MANIFEST_OUTPUT_DIR/openshift/alm" \
    "$DEPLOY_MANIFESTS_DIR/common-alm-values.yaml" \
    "$DEPLOY_MANIFESTS_DIR/openshift-alm-values.yaml" \
    "$ALM_OVERRIDE_VALUES_FILE"

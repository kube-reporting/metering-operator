#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are unspecified, then this script
# requires 1 argument, the image tag for the helm-operator pod. Additionally,
# a second optional argument can be provided to override where the manifests
# should be output, by default it outputs everything into
# deploy/manifests/{generic,openshift,tectonic} directories.
#
# If both of those environment variables are set, then this script takes no
# arguments and the environment variables will be used as the paths to files
# containing override values for rendering the manifests to the output
# directory. In this scenario, a first optional argument can be provided to
# control the manifests output directory.

set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

TMPDIR="$(mktemp -d)"

IMAGE_TAG=""

if [[ -z "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES" && -z "$CUSTOM_ALM_OVERRIDE_VALUES" ]]; then
    : "${1?"Usage: $0 IMAGE_TAG"}"
    echo "Using $1 as IMAGE_TAG"
    IMAGE_TAG="$1"
    shift
fi

DEPLOY_DIR="$DIR/../manifests/deploy"
# By default, we output into the deploy directory, but this can be overridden
# by passing a second argument to this script
OUTPUT_DIR="${1:-$DEPLOY_DIR}"

echo "Using $OUTPUT_DIR as output directory"
echo

trap "rm -rf $TMPDIR" EXIT

HELM_OPERATOR_OVERRIDE_VALUES_FILE=""
ALM_OVERRIDE_VALUES_FILE=""

if [ -n "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES" ]; then
    HELM_OPERATOR_OVERRIDE_VALUES_FILE="$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
else
    HELM_OPERATOR_OVERRIDE_VALUES_FILE="$TMPDIR/override-helm-operator-values.yaml"
    "$DIR/render-helm-operator-override-values.sh" "$IMAGE_TAG" > "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"
fi

if [ -n "$CUSTOM_ALM_OVERRIDE_VALUES" ]; then
    ALM_OVERRIDE_VALUES_FILE="$CUSTOM_ALM_OVERRIDE_VALUES"
else
    ALM_OVERRIDE_VALUES_FILE="$TMPDIR/override-alm-values.yaml"
    "$DIR/render-alm-override-values.sh" "$IMAGE_TAG" > "$ALM_OVERRIDE_VALUES_FILE"
fi

# tectonic
echo "Creating Tectonic deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/tectonic/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/tectonic-helm-operator-values.yaml" \
    "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"

echo
echo "Creating Tectonic ALM manifests"
"$DIR/create-alm-manifests.sh" \
    "$OUTPUT_DIR/tectonic/helm-operator" \
    "$OUTPUT_DIR/tectonic/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/tectonic-alm-values.yaml" \
    "$ALM_OVERRIDE_VALUES_FILE"

# openshift
echo
echo "Creating Openshift deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/openshift/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/openshift-helm-operator-values.yaml" \
    "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"

echo
echo "Creating Openshift ALM manifests"
"$DIR/create-alm-manifests.sh" \
    "$OUTPUT_DIR/openshift/helm-operator" \
    "$OUTPUT_DIR/openshift/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/openshift-alm-values.yaml" \
    "$ALM_OVERRIDE_VALUES_FILE"

# generic
echo
echo "Creating Generic deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/generic/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/generic-helm-operator-values.yaml" \
    "$HELM_OPERATOR_OVERRIDE_VALUES_FILE"

echo
echo "Creating Generic ALM manifests"
"$DIR/create-alm-manifests.sh" \
    "$OUTPUT_DIR/generic/helm-operator" \
    "$OUTPUT_DIR/generic/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/generic-alm-values.yaml" \
    "$ALM_OVERRIDE_VALUES_FILE"

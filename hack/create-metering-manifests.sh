#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${1?"Usage: $0 IMAGE_TAG"}"

TMPDIR="$(mktemp -d)"
DEPLOY_DIR="$DIR/../manifests/deploy"
# By default, we output into the deploy directory, but this can be overridden
OUTPUT_DIR="${DEPLOY_DIR:-$2}"

echo "Using $DEPLOY_DIR as output directory"

trap "rm -rf $TMPDIR" EXIT

"$DIR/render-helm-operator-override-values.sh" "$1" > "$TMPDIR/override-helm-operator-values.yaml"
"$DIR/render-alm-override-values.sh" "$1" > "$TMPDIR/override-alm-values.yaml"

# tectonic
echo "Creating Tectonic deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/tectonic/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/tectonic-helm-operator-values.yaml" \
    "$TMPDIR/override-helm-operator-values.yaml"

echo
echo "Creating Tectonic ALM manifests"
"$DIR/create-alm-csv-manifests.sh" \
    "$OUTPUT_DIR/tectonic/helm-operator" \
    "$OUTPUT_DIR/tectonic/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/tectonic-alm-values.yaml" \
    "$TMPDIR/override-alm-values.yaml"

# openshift
echo
echo "Creating Openshift deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/openshift/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/openshift-helm-operator-values.yaml" \
    "$TMPDIR/override-helm-operator-values.yaml"

echo
echo "Creating Openshift ALM manifests"
"$DIR/create-alm-csv-manifests.sh" \
    "$OUTPUT_DIR/openshift/helm-operator" \
    "$OUTPUT_DIR/openshift/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/openshift-alm-values.yaml" \
    "$TMPDIR/override-alm-values.yaml"

# generic
echo
echo "Creating Generic deploy manifests"
"$DIR/create-deploy-manifests.sh" \
    "$OUTPUT_DIR/generic/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/generic-helm-operator-values.yaml" \
    "$TMPDIR/override-helm-operator-values.yaml"

echo
echo "Creating Generic ALM manifests"
"$DIR/create-alm-csv-manifests.sh" \
    "$OUTPUT_DIR/generic/helm-operator" \
    "$OUTPUT_DIR/generic/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$TMPDIR/override-alm-values.yaml"

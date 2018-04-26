#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${1?"Usage: $0 IMAGE_TAG"}"

TMPDIR="$(mktemp -d)"
DEPLOY_DIR="$DIR/../manifests/deploy"

trap "rm -rf $TMPDIR" EXIT

"$DIR/render-helm-operator-override-values.sh" "$1" > "$TMPDIR/override-helm-operator-values.yaml"
"$DIR/render-alm-override-values.sh" "$1" > "$TMPDIR/override-alm-values.yaml"

"$DIR/create-deploy-manifests.sh" \
    "$DEPLOY_DIR/tectonic/helm-operator" \
    "$DEPLOY_DIR/common-helm-operator-values.yaml" \
    "$DEPLOY_DIR/tectonic/helm-operator-values.yaml" \
    "$TMPDIR/override-helm-operator-values.yaml"

"$DIR/create-alm-csv-manifests.sh" \
    "$DEPLOY_DIR/tectonic/helm-operator" \
    "$DEPLOY_DIR/tectonic/alm" \
    "$DEPLOY_DIR/common-alm-values.yaml" \
    "$DEPLOY_DIR/tectonic/alm-values.yaml" \
    "$TMPDIR/override-alm-values.yaml"


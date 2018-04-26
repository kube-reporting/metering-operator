#!/bin/bash -e

set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/helm-operator"
: "${MANIFESTS_DIR:=$DIR/../manifests}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions}"

if [[ $# -ge 2 ]] ; then
    OUTPUT_DIR=$1
    echo "Output directory: ${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
    shift

    echo "Values ***REMOVED***les: $*"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
***REMOVED***
    echo "Must specify output directory and values ***REMOVED***les"
    exit 1
***REMOVED***

echo "CRD manifest directory: $CRD_DIR"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/role.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$OUTPUT_DIR/metering-helm-operator-role.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/rolebinding.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$OUTPUT_DIR/metering-helm-operator-rolebinding.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/deployment.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$OUTPUT_DIR/metering-helm-operator-deployment.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/service-account.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$OUTPUT_DIR/metering-helm-operator-service-account.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/crd.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$CRD_DIR/metering.crd.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/cr.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$OUTPUT_DIR/metering.yaml"

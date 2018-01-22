#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/helm-operator"
VALUES_ARGS=()

if [[ $# -eq 0 ]] ; then
    echo "No arguments provided, using chargeback-helm-operator-values.yaml"
    VALUES_ARGS=(-f "$DIR/chargeback-helm-operator-values.yaml")
***REMOVED***
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
***REMOVED***

: "${INSTALLER_MANIFEST_DIR:=$DIR/../manifests/installer}"

echo "Installer manifest directory: $INSTALLER_MANIFEST_DIR"

mkdir -p "${INSTALLER_MANIFEST_DIR}"

helm template "$CHART" "${VALUES_ARGS[@]}" -x "templates/rbac.yaml" > \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-rbac.yaml"
helm template "$CHART" "${VALUES_ARGS[@]}" -x "templates/deployment.yaml" > \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml"
helm template "$CHART" "${VALUES_ARGS[@]}" -x "templates/crd.yaml" > \
    "$INSTALLER_MANIFEST_DIR/chargeback-crd.yaml"
helm template "$CHART" "${VALUES_ARGS[@]}" -x "templates/service-account.yaml" > \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml"


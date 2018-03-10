#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/helm-operator"

# We use cd + pwd in a subshell to turn this into an absolute path (readlink -f isn't cross platform)
OUTPUT_DIR="$(cd "${OUTPUT_DIR:=$DIR/..}" && pwd)"
: "${MANIFESTS_DIR:=$OUTPUT_DIR/manifests}"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/installer}"

: "${HELM_OPERATOR_VALUES_FILE:=$DIR/chargeback-helm-operator-values.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-definitions}"


echo "helm-operator values file: $HELM_OPERATOR_VALUES_FILE"
VALUES_ARGS=(-f "$HELM_OPERATOR_VALUES_FILE")

if [[ $# -ne 0 ]] ; then
    echo "Extra values files: $*"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
fi
echo "Output directory: ${OUTPUT_DIR}"
echo "Installer manifest directory: $INSTALLER_MANIFEST_DIR"
echo "CRD manifest directory: $CRD_DIR"

mkdir -p "${INSTALLER_MANIFEST_DIR}" "${CRD_DIR}"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/role.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-role.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/rolebinding.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-rolebinding.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/deployment.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/service-account.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/crd.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$CRD_DIR/chargeback.crd.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/cr.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback.yaml"

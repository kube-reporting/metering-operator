#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/helm-operator"

# We use cd + pwd in a subshell to turn this into an absolute path (readlink -f isn't cross platform)
OUTPUT_DIR="$(cd "${OUTPUT_DIR:=$DIR/..}" && pwd)"
: "${MANIFESTS_DIR:=$OUTPUT_DIR/manifests}"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/installer}"

: "${HELM_OPERATOR_VALUES_FILE:=$DIR/chargeback-helm-operator-values.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions}"

echo "helm-operator values ***REMOVED***le: $HELM_OPERATOR_VALUES_FILE"
echo "Output directory: ${OUTPUT_DIR}"
echo "Installer manifest directory: $INSTALLER_MANIFEST_DIR"
echo "CRD manifest directory: $CRD_DIR"

mkdir -p "${INSTALLER_MANIFEST_DIR}" "${CRD_DIR}"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/role.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-role.yaml"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/rolebinding.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-rolebinding.yaml"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/deployment.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/service-account.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/crd.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$CRD_DIR/chargeback.crd.yaml"

helm template "$CHART" \
    -f "$HELM_OPERATOR_VALUES_FILE" \
    -x "templates/cr.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$INSTALLER_MANIFEST_DIR/chargeback.yaml"

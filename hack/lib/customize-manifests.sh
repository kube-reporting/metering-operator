#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/../..
source "${ROOT_DIR}/hack/common.sh"

customizeMeteringOperatorDeployment() {
    OUTPUT_DIR="${1:?}"
    if [[ -n "${METERING_OPERATOR_IMAGE_REPO:-}" && -n "${METERING_OPERATOR_IMAGE_TAG:-}" ]]; then
        echo "using \$METERING_OPERATOR_IMAGE_REPO=$METERING_OPERATOR_IMAGE_REPO to override metering-operator image"
        echo "using \$METERING_OPERATOR_IMAGE_TAG=$METERING_OPERATOR_IMAGE_TAG to override metering-operator image tag"
    ***REMOVED***
    if [[ -n "${METERING_OPERATOR_ALL_NAMESPACES:-}" ]]; then
        echo "using \$METERING_OPERATOR_ALL_NAMESPACES=$METERING_OPERATOR_ALL_NAMESPACES"
    ***REMOVED***
    if [[ -n "${METERING_OPERATOR_TARGET_NAMESPACES:-}" ]]; then
        echo "using \$METERING_OPERATOR_TARGET_NAMESPACES=$METERING_OPERATOR_TARGET_NAMESPACES"
    ***REMOVED***

    export METERING_NAMESPACE METERING_OPERATOR_IMAGE_REPO METERING_OPERATOR_IMAGE_TAG METERING_OPERATOR_ALL_NAMESPACES METERING_OPERATOR_TARGET_NAMESPACES
    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        -F "$ROOT_DIR/hack/jq/custom-metering-operator-deployment.jq" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-deployment.yaml" \
        > "$OUTPUT_DIR/metering-operator-deployment.yaml"
}

customizeMeteringOperatorRolebinding(){
    OUTPUT_DIR="${1:?}"
    # create a role and rolebinding in each namespace

    export METERING_NAMESPACE METERING_OPERATOR_TARGET_NAMESPACES

    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        -F "$ROOT_DIR/hack/jq/custom-metering-rolebinding.jq" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml" \
        > "$OUTPUT_DIR/metering-operator-rolebinding.yaml"

    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        -F "$ROOT_DIR/hack/jq/custom-metering-role.jq" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
        > "$OUTPUT_DIR/metering-operator-role.yaml"
}

customizeMeteringOperatorClusterRolebinding() {
    OUTPUT_DIR="${1:?}"
    # to set the ServiceAccount subject namespace, since it's cluster
    # scoped.  updating the name is to avoid conflicting with others also
    # using this script to install.

    export METERING_NAMESPACE
    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        '.metadata.name=$ENV.METERING_NAMESPACE + "-" + .metadata.name | .subjects[0].namespace=$ENV.METERING_NAMESPACE | .roleRef.name=.metadata.name' \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrolebinding.yaml" \
        > "$OUTPUT_DIR/metering-operator-clusterrolebinding.yaml"

    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        '.metadata.name=$ENV.METERING_NAMESPACE + "-" + .metadata.name' \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrole.yaml" \
        > "$OUTPUT_DIR/metering-operator-clusterrole.yaml"
}


customizeMeteringInstallManifests() {
    OUTPUT_DIR="${1:?}"
    customizeMeteringOperatorDeployment "$OUTPUT_DIR"

    if [ -n "${METERING_OPERATOR_TARGET_NAMESPACES:-}" ]; then
        echo "using \$METERING_OPERATOR_TARGET_NAMESPACES=$METERING_OPERATOR_TARGET_NAMESPACES as target namespaces for metering-operator"
    ***REMOVED***
    customizeMeteringOperatorRolebinding "$OUTPUT_DIR"

    if [[ "${METERING_INSTALL_CLUSTERROLEBINDING}" == "true" || "${METERING_UNINSTALL_CLUSTERROLEBINDING}" == "true" ]]; then
        echo "Updating Metering ClusterRole and ClusterRoleBinding"
    ***REMOVED***
    customizeMeteringOperatorClusterRolebinding "$OUTPUT_DIR"
}

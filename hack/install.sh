#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

# can also be speci***REMOVED***ed as an argument
METERING_CR_FILE="${1:-$METERING_CR_FILE}"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${METERING_NAMESPACE}"
    kubectl create namespace "${METERING_NAMESPACE}" || true
elif ! kubectl get namespace "${METERING_NAMESPACE}" 2> /dev/null; then
    echo "Namespace '${METERING_NAMESPACE}' does not exist, please create it before starting"
    exit 1
***REMOVED***

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"

if [ "$SKIP_METERING_OPERATOR_DEPLOYMENT" == "true" ]; then
    echo "\$SKIP_METERING_OPERATOR_DEPLOYMENT=true, not creating metering-operator"
***REMOVED***
    msg "Installing metering-operator service account and RBAC resources"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml"

    if [ "${METERING_INSTALL_REPORTING_OPERATOR_CLUSTERROLEBINDING}" == "true" ]; then
        msg "Installing metering-operator Cluster level RBAC resources"

        TMPDIR="$(mktemp -d)"
        trap "rm -rf $TMPDIR" EXIT

        # to set the ServiceAccount subject namespace, since it's cluster
        # scoped.  updating the name is to avoid conflicting with others also
        # using this script to install.

        "$ROOT_DIR/hack/yamltojson" < "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrolebinding.yaml" \
            | jq -r '.metadata.name=$namespace + "-" + .metadata.name | .subjects[0].namespace=$namespace | .roleRef.name=.metadata.name' \
            --arg namespace "$METERING_NAMESPACE" \
            > "$TMPDIR/metering-operator-clusterrolebinding.yaml"

        "$ROOT_DIR/hack/yamltojson" < "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrole.yaml" \
            | jq -r '.metadata.name=$namespace + "-" + .metadata.name' \
            --arg namespace "$METERING_NAMESPACE" \
            > "$TMPDIR/metering-operator-clusterrole.yaml"

        kube-install \
            "$TMPDIR/metering-operator-clusterrole.yaml" \
            "$TMPDIR/metering-operator-clusterrolebinding.yaml"
    ***REMOVED***

    msg "Installing metering-operator"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-deployment.yaml"
***REMOVED***

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

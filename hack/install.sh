#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/customize-manifests.sh"

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
***REMOVED***nd "$CRD_DIR" -type f -name '*.crd.yaml' -exec kubectl apply -f {} \;

if [ "$SKIP_METERING_OPERATOR_DEPLOYMENT" == "true" ]; then
    echo "\$SKIP_METERING_OPERATOR_METERING_OPERATOR_DEPLOYMENT=true, not creating metering-operator"
***REMOVED***
    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT SIGINT

    cp -r "$INSTALLER_MANIFESTS_DIR" "$TMPDIR"
    customizeMeteringInstallManifests "$TMPDIR"

    msg "Installing metering-operator service account and RBAC resources"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"

    kubectl apply \
        -f "$TMPDIR/metering-operator-rolebinding.yaml" \
        -f "$TMPDIR/metering-operator-role.yaml"

    if [ "${METERING_INSTALL_CLUSTERROLEBINDING}" == "true" ]; then
        msg "Installing metering-operator Cluster level RBAC resources"

        kubectl apply \
        -f "$TMPDIR/metering-operator-clusterrole.yaml" \
        -f "$TMPDIR/metering-operator-clusterrolebinding.yaml"
    ***REMOVED***

    msg "Installing metering-operator"
    kube-install \
        "$TMPDIR/metering-operator-deployment.yaml"
***REMOVED***

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

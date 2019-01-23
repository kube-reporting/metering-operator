#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

# can also be specified as an argument
METERING_CR_FILE="${1:-$METERING_CR_FILE}"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${METERING_NAMESPACE}"
    kubectl create namespace "${METERING_NAMESPACE}" || true
elif ! kubectl get namespace "${METERING_NAMESPACE}" 2> /dev/null; then
    echo "Namespace '${METERING_NAMESPACE}' does not exist, please create it before starting"
    exit 1
fi

msg "Installing Custom Resource Definitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-definitions"

if [ "$SKIP_METERING_OPERATOR_DEPLOYMENT" == "true" ]; then
    echo "\$SKIP_METERING_OPERATOR_DEPLOYMENT=true, not creating metering-operator"
else
    msg "Installing metering-operator service account and RBAC resources"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"

    TMPDIR="$(mktemp -d)"
    trap "rm -rf $TMPDIR" EXIT

    # if $METERING_OPERATOR_TARGET_NAMESPACES is set, then install the
    # metering-operator role and rolebinding in each namespace configured to
    # grant the metering-operator serviceAccount permissions
    if [ -z "${METERING_OPERATOR_TARGET_NAMESPACES:-}" ]; then
        kube-install \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml"
    else
        while read -rd, TARGET_NS; do
            "$ROOT_DIR/hack/yamltojson" < "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml" \
                | jq -r '.subjects[0].namespace=$namespace' \
                --arg namespace "$METERING_NAMESPACE" \
                > "$TMPDIR/metering-operator-rolebinding.yaml"

            # the role is unmodified
            kubectl apply -n "$TARGET_NS" -f "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml"
            kubectl apply -n "$TARGET_NS" -f "$TMPDIR/metering-operator-rolebinding.yaml"

        done <<<"$METERING_OPERATOR_TARGET_NAMESPACES,"
    fi

    if [ "${METERING_INSTALL_REPORTING_OPERATOR_CLUSTERROLEBINDING}" == "true" ]; then
        msg "Installing metering-operator Cluster level RBAC resources"

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
    fi

    msg "Installing metering-operator"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-deployment.yaml"
fi

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

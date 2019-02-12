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
    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT SIGINT

    if [ "$USE_CUSTOM_METERING_OPERATOR" == "true" ]; then
        echo "\$USE_CUSTOM_METERING_OPERATOR=true, using custom metering-operator configuration"

        export METERING_OPERATOR_IMAGE="${CUSTOM_METERING_OPERATOR_IMAGE:?}"
        export METERING_OPERATOR_IMAGE_TAG="${CUSTOM_METERING_OPERATOR_IMAGE_TAG:?}"
        echo "using \$CUSTOM_METERING_OPERATOR_IMAGE=$CUSTOM_METERING_OPERATOR_IMAGE and \$CUSTOM_METERING_OPERATOR_IMAGE_TAG=$CUSTOM_METERING_OPERATOR_IMAGE_TAG to override metering-operator image"

        # render out custom helm operator override values if CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES isn't set
        if [ -z "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-}" ]; then
            export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES="$TMPDIR/override-helm-operator-values.yaml"
            "$ROOT_DIR/hack/render-helm-operator-override-values.sh" > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
        fi

        # render out custom alm override values if CUSTOM_ALM_OVERRIDE_VALUES isn't set
        if [ -z "${CUSTOM_ALM_OVERRIDE_VALUES:-}" ]; then
            export CUSTOM_ALM_OVERRIDE_VALUES="$TMPDIR/override-alm-values.yaml"
            "$ROOT_DIR/hack/render-alm-override-values.sh" > "$CUSTOM_ALM_OVERRIDE_VALUES"
        fi

        export MANIFEST_OUTPUT_DIR="$TMPDIR"
        "$ROOT_DIR/hack/create-metering-manifests.sh"

        # override DEPLOY_MANIFESTS_DIR since we've modified the files
        export DEPLOY_MANIFESTS_DIR="$TMPDIR"
        # update INSTALLER_MANIFESTS_DIR used below to use new DEPLOY_MANIFESTS_DIR
        export INSTALLER_MANIFESTS_DIR="$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator"
    fi

    msg "Installing metering-operator service account and RBAC resources"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"

    # if $METERING_OPERATOR_TARGET_NAMESPACES is set, then install the
    # metering-operator role and rolebinding in each namespace configured to
    # grant the metering-operator serviceAccount permissions
    if [ -z "${METERING_OPERATOR_TARGET_NAMESPACES:-}" ]; then
        kube-install \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml"
    else
        while read -rd, TARGET_NS; do
            # shellcheck disable=SC2016
            "$FAQ_BIN" -f yaml -o yaml -M -c -r \
                --kwargs "namespace=$METERING_NAMESPACE" \
                '.subjects[0].namespace=$namespace' \
                "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml" \
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

        # shellcheck disable=SC2016
        "$FAQ_BIN" -f yaml -o yaml -M -c -r \
            --kwargs "namespace=$METERING_NAMESPACE" \
            '.metadata.name=$namespace + "-" + .metadata.name | .subjects[0].namespace=$namespace | .roleRef.name=.metadata.name' \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrolebinding.yaml" \
            > "$TMPDIR/metering-operator-clusterrolebinding.yaml"

        # shellcheck disable=SC2016
        "$FAQ_BIN" -f yaml -o yaml -M -c -r \
            --kwargs "namespace=$METERING_NAMESPACE" \
            '.metadata.name=$namespace + "-" + .metadata.name' \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrole.yaml" \
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

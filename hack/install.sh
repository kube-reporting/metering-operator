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
    TMPDIR="$(mktemp -d)"
    trap "rm -rf $TMPDIR" EXIT SIGINT

    if [ "$USE_CUSTOM_METERING_OPERATOR_IMAGE" == "true" ]; then
        echo "\$USE_CUSTOM_METERING_OPERATOR_IMAGE=true, using \$CUSTOM_METERING_OPERATOR_IMAGE and \$CUSTOM_METERING_OPERATOR_IMAGE_TAG to override metering-operator image"
        export METERING_OPERATOR_IMAGE="${CUSTOM_METERING_OPERATOR_IMAGE:?}"
        export METERING_OPERATOR_IMAGE_TAG="${CUSTOM_METERING_OPERATOR_IMAGE_TAG:?}"

        # render out custom helm operator override values if CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES isn't set
        if [ -z "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-}" ]; then
            export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES="$TMPDIR/override-helm-operator-values.yaml"
            "$ROOT_DIR/hack/render-helm-operator-override-values.sh" > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
        ***REMOVED***

        # render out custom alm override values if CUSTOM_ALM_OVERRIDE_VALUES isn't set
        if [ -z "${CUSTOM_ALM_OVERRIDE_VALUES:-}" ]; then
            export CUSTOM_ALM_OVERRIDE_VALUES="$TMPDIR/override-alm-values.yaml"
            "$ROOT_DIR/hack/render-alm-override-values.sh" > "$CUSTOM_ALM_OVERRIDE_VALUES"
        ***REMOVED***

        export MANIFEST_OUTPUT_DIR="$TMPDIR"
        "$ROOT_DIR/hack/create-metering-manifests.sh"

        # override DEPLOY_MANIFESTS_DIR since we've modi***REMOVED***ed the ***REMOVED***les
        export DEPLOY_MANIFESTS_DIR="$TMPDIR"
        # update INSTALLER_MANIFESTS_DIR used below to use new DEPLOY_MANIFESTS_DIR
        export INSTALLER_MANIFESTS_DIR="$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator"
    ***REMOVED***

    msg "Installing metering-operator service account and RBAC resources"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"

    TMPDIR="$(mktemp -d)"
    trap "rm -rf $TMPDIR" EXIT

    # if $METERING_OPERATOR_TARGET_NAMESPACES is set, then install the
    # metering-operator role and rolebinding in each namespace con***REMOVED***gured to
    # grant the metering-operator serviceAccount permissions
    if [ -z "${METERING_OPERATOR_TARGET_NAMESPACES:-}" ]; then
        kube-install \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
            "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml"
    ***REMOVED***
        while read -rd, TARGET_NS; do
            "$ROOT_DIR/hack/yamltojson" < "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml" \
                | jq -r '.subjects[0].namespace=$namespace' \
                --arg namespace "$METERING_NAMESPACE" \
                > "$TMPDIR/metering-operator-rolebinding.yaml"

            # the role is unmodi***REMOVED***ed
            kubectl apply -n "$TARGET_NS" -f "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml"
            kubectl apply -n "$TARGET_NS" -f "$TMPDIR/metering-operator-rolebinding.yaml"

        done <<<"$METERING_OPERATOR_TARGET_NAMESPACES,"
    ***REMOVED***

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
    ***REMOVED***

    msg "Installing metering-operator"
    kube-install \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-deployment.yaml"
***REMOVED***

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

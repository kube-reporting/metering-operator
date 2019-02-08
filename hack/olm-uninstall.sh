#!/bin/bash -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

set +e

if [ "$METERING_NAMESPACE" != "openshift-metering" ]; then
    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT SIGINT

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.targetNamespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.catalogsourcecon***REMOVED***g.yaml" \
        > "$TMPDIR/metering.catalogsourcecon***REMOVED***g.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.targetNamespaces[0]=$namespace | .metadata.name=$namespace + "-" + .metadata.name' \
        "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml" \
        > "$TMPDIR/metering.operatorgroup.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.sourceNamespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.subscription.yaml" \
        > "$TMPDIR/metering.subscription.yaml"

        export OLM_MANIFESTS_DIR="$TMPDIR"
***REMOVED***

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing Metering Subscription"
kube-remove \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Removing Metering Operator Group"
kube-remove \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Removing Metering Catalog Source Con***REMOVED***g"
kubectl delete -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsourcecon***REMOVED***g.yaml"

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
        "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"
***REMOVED***

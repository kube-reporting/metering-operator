#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/customize-manifests.sh"

set +e

TMPDIR="$(mktemp -d)"
# shellcheck disable=SC2064
trap "rm -rf $TMPDIR" EXIT SIGINT

cp -r "$INSTALLER_MANIFESTS_DIR" "$TMPDIR"
customizeMeteringInstallManifests "$TMPDIR"


msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing metering-operator"
kube-remove \
    "$TMPDIR/metering-operator-deployment.yaml"

msg "Removing metering-operator service account and RBAC resources"
kubectl delete \
    -f "$TMPDIR/metering-operator-rolebinding.yaml" \
    -f "$TMPDIR/metering-operator-role.yaml"

kube-remove \
    "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"

if [ "${METERING_UNINSTALL_CLUSTERROLEBINDING}" == "true" ]; then
    msg "Removing metering-operator Cluster level RBAC resources"

    kubectl delete \
        -f "$TMPDIR/metering-operator-clusterrole.yaml" \
        -f "$TMPDIR/metering-operator-clusterrolebinding.yaml"
***REMOVED***


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
        "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"
***REMOVED***

if [ "$DELETE_PVCS" == "true" ]; then
    echo "Deleting PVCs"
    kube-remove-non-***REMOVED***le pvc -l "app in (hive-metastore, hdfs-namenode, hdfs-datanode)"
***REMOVED***

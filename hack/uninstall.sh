#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

set +e

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing metering-operator"
kube-remove \
    "$INSTALLER_MANIFESTS_DIR/metering-operator-deployment.yaml"

msg "Removing metering-operator service account and RBAC resources"
kube-remove \
    "$INSTALLER_MANIFESTS_DIR/metering-operator-rolebinding.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-operator-role.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-operator-service-account.yaml"


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource Definitions"
else
    msg "Removing Custom Resource Definitions"
    kube-remove \
        manifests/custom-resource-definitions
fi

if [ "$DELETE_PVCS" == "true" ]; then
    echo "Deleting PVCs"
    kube-remove-non-file pvc -l "app in (hive-metastore, hdfs-namenode, hdfs-datanode)"
fi

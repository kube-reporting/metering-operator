#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

MANIFESTS_DIR="$DIR/../manifests"
: "${DEPLOY_PLATFORM:=generic}"
: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"
: "${DELETE_PVCS:=false}"
: "${SKIP_DELETE_CRDS:=true}"

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing metering-helm-operator"
kube-remove \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-deployment.yaml"

msg "Removing metering-helm-operator service account and RBAC resources"
kube-remove \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-rolebinding.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-role.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-service-account.yaml"


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

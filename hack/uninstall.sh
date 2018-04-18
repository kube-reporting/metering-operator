#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

: "${INSTALLER_MANIFEST_DIR:=$DIR/../manifests/installer}"
: "${CHARGEBACK_CR_FILE:=$INSTALLER_MANIFEST_DIR/chargeback.yaml}"
: "${DELETE_PVCS:=false}"
: "${SKIP_DELETE_CRDS:=true}"

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Removing pull secrets"
    kube-remove-non-file secret coreos-pull-secret
fi

msg "Removing Chargeback Resource"
kube-remove \
    "$CHARGEBACK_CR_FILE"

msg "Removing chargeback-helm-operator"
kube-remove \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml"

msg "Removing chargeback-helm-operator service account and RBAC resources"
kube-remove \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-rolebinding.yaml" \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-role.yaml" \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml"


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

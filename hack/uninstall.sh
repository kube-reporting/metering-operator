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


if [ "${METERING_UNINSTALL_REPORTING_OPERATOR_CLUSTERROLEBINDING}" == "true" ]; then
    msg "Removing metering-operator Cluster level RBAC resources"

    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT

    # we modify the name of these resources to be unique by namespace, and
    # to set the ServiceAccount subject namespace, since it's cluster
    # scoped.  updating the name is to avoid conflicting with others also
    # using this script to install.
    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r '.metadata.name=$namespace + "-" + .metadata.name | .subjects[0].namespace=$namespace | .roleRef.name=.metadata.name' \
        --arg namespace "$METERING_NAMESPACE" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrolebinding.yaml" \
        > "$TMPDIR/metering-operator-clusterrolebinding.yaml"

    # shellcheck disable=SC2016
    "$FAQ_BIN" -f yaml -o yaml -M -c -r '.metadata.name=$namespace + "-" + .metadata.name' \
        --arg namespace "$METERING_NAMESPACE" \
        "$INSTALLER_MANIFESTS_DIR/metering-operator-clusterrole.yaml" \
        > "$TMPDIR/metering-operator-clusterrole.yaml"

    kube-remove \
        "$TMPDIR/metering-operator-clusterrole.yaml" \
        "$TMPDIR/metering-operator-clusterrolebinding.yaml"
***REMOVED***


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
        manifests/custom-resource-de***REMOVED***nitions
***REMOVED***

if [ "$DELETE_PVCS" == "true" ]; then
    echo "Deleting PVCs"
    kube-remove-non-***REMOVED***le pvc -l "app in (hive-metastore, hdfs-namenode, hdfs-datanode)"
***REMOVED***

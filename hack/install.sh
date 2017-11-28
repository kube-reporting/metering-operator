#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

: "${CREATE_NAMESPACE:=false}"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${CHARGEBACK_NAMESPACE}"
    kubectl create namespace "${CHARGEBACK_NAMESPACE}" || true
elif ! kubectl get namespace ${CHARGEBACK_NAMESPACE} 2> /dev/null; then
    echo "Namespace '${CHARGEBACK_NAMESPACE}' does not exist, please create it before starting"
    exit 1
fi

msg "Configuring pull secrets"
copy-tectonic-pull

msg "Installing Custom Resource Definitions"
kube-install \
    manifests/custom-resource-definitions

msg "Installing chargeback-alm-install service account and RBAC resources"
kube-install \
    manifests/installer/chargeback-alm-install-service-account.yaml \
    manifests/installer/chargeback-alm-install-rbac.yaml

msg "Installing alm-install-operator"
kube-install \
    manifests/installer/chargeback-alm-install-deployment.yaml


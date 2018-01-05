#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

kubectl create namespace "${CHARGEBACK_NAMESPACE}" || true

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Configuring pull secrets"
    copy-tectonic-pull
fi

msg "Installing Custom Resource Definitions"
kube-install \
    manifests/custom-resource-definitions

msg "Installing Chargeback Cluster Service Version"
kube-install \
    manifests/alm/chargeback-clusterserviceversion.yaml

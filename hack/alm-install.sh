#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

msg "Configuring pull secrets"
copy-tectonic-pull

msg "Installing Custom Resource Definitions"
kube-install \
    manifests/custom-resource-definitions

msg "Installing Chargeback Cluster Service Version"
kube-install \
    manifests/alm/chargeback-clusterserviceversion.yaml

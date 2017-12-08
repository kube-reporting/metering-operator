#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Con***REMOVED***guring pull secrets"
copy-tectonic-pull

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    manifests/custom-resource-de***REMOVED***nitions

msg "Installing Chargeback Cluster Service Version"
kube-install \
    manifests/alm/chargeback-clusterserviceversion.yaml

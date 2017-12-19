#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Con***REMOVED***guring pull secrets"
    copy-tectonic-pull
***REMOVED***

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    manifests/custom-resource-de***REMOVED***nitions

msg "Installing Chargeback Cluster Service Version"
kube-install \
    manifests/alm/chargeback-clusterserviceversion.yaml

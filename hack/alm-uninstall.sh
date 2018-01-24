#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Removing pull secrets"
    kube-remove-non-***REMOVED***le secret coreos-pull-secret
***REMOVED***

msg "Removing Chargeback Resource"
kube-remove \
    manifests/installer/chargeback.yaml

msg "Removing Chargeback Cluster Service Version"
kube-remove \
    manifests/alm/chargeback.clusterserviceversion.yaml

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
        manifests/custom-resource-de***REMOVED***nitions
***REMOVED***

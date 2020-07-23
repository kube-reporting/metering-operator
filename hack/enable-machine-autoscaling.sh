#! /bin/bash

set -eoux pipefail

: "${KUBECONFIG:?}"
ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
MANIFESTS_BASE_PATH="${ROOT_DIR}/olm_deploy/manifests/machines"

if ! kubectl get clusterautoscalers default > /dev/null 2>&1; then
    kubectl create -f ${MANIFESTS_BASE_PATH}/clusterautoscaler.yaml
fi

MACHINESETS=( $(oc -n openshift-machine-api get machinesets --no-headers | awk '{ print $1 }') )
for machine in "${MACHINESETS[@]}"; do
    export MACHINE_NAME=$(oc -n openshift-machine-api get machineset $machine -o jsonpath='{.metadata.name}')
    envsubst < ${MANIFESTS_BASE_PATH}/machineautoscaler.yaml | kubectl apply -f -
done

#!/usr/bin/env bash

set -e

export CHARGEBACK_NAMESPACE="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
export PRIVILEGED_INSTALL="false"

CHARGEBACK_DEPLOY="$(kubectl get deploy -n $CHARGEBACK_NAMESPACE -l app=chargeback -o json | jq '.items[]')"

if [[ $CHARGEBACK_DEPLOY == "" ]]; then
    /opt/hack/install.sh
else
    echo "chargeback deployment exists, skipping installation"
fi

while true; do
    # we don't want this to exit once things are installed
    sleep 600
done

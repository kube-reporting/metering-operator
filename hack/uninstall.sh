#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing chargeback namespace"
kube-remove manifests/chargeback/namespace.yaml

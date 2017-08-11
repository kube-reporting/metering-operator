#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
. ${DIR}/util.sh

msg "Removing chargeback namespace"
kube-remove manifests/chargeback/namespace.yaml

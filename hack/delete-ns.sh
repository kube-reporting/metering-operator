#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export CHARGEBACK_NAMESPACE=${CHARGEBACK_NAMESPACE:-chargeback-ci}
CHARGEBACK_NAMESPACE="$(sanetize_namespace "$CHARGEBACK_NAMESPACE")"

kubectl delete ns --now --ignore-not-found=true "$CHARGEBACK_NAMESPACE"

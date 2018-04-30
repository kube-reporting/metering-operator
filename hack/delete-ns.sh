#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci}
METERING_NAMESPACE="$(sanetize_namespace "$METERING_NAMESPACE")"

kubectl delete ns --now --ignore-not-found=true "$METERING_NAMESPACE"

#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
name: metering-helm-operator.v$1
spec:
  version: $1
  labels:
    alm-status-descriptors: metering-helm-operator.v$1
    alm-owner-metering: metering-helm-operator
  matchLabels:
    alm-owner-metering: metering-helm-operator
EOF

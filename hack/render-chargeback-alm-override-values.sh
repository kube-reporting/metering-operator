#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
name: chargeback-helm-operator.v$1
spec:
  version: $1
  labels:
    alm-status-descriptors: chargeback-helm-operator.v$1
    alm-owner-chargeback: chargeback-helm-operator
  matchLabels:
    alm-owner-chargeback: chargeback-helm-operator
EOF

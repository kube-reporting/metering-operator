#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
name: metering-operator.v$1
spec:
  version: $1
  labels:
    alm-status-descriptors: metering-operator.v$1
    alm-owner-metering: metering-operator
  matchLabels:
    alm-owner-metering: metering-operator
channels:
- name: alpha
  currentCSV: metering-operator.v$1
EOF

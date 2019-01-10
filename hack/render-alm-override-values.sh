#!/bin/bash

set -e
set -u

cat <<EOF
name: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
spec:
  version: ${METERING_OPERATOR_IMAGE_TAG}
  labels:
    alm-status-descriptors: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
    alm-owner-metering: metering-operator
  matchLabels:
    alm-owner-metering: metering-operator
channels:
- name: alpha
  currentCSV: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
EOF

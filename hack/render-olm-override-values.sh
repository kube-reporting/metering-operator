#!/bin/bash

set -e
set -u

cat <<EOF
csv:
  name: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
  version: ${METERING_OPERATOR_IMAGE_TAG}
  labels:
    olm-status-descriptors: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
    olm-owner-metering: metering-operator
  matchLabels:
    olm-owner-metering: metering-operator
channels:
- name: alpha
  currentCSV: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
EOF

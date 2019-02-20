#!/bin/bash

set -e
set -u

cat <<EOF
csv:
  name: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
  version: ${METERING_OPERATOR_IMAGE_TAG}
annotations:
  containerImage: quay.io/coreos/metering-helm-operator:${METERING_OPERATOR_IMAGE_TAG}
channels:
- name: alpha
  currentCSV: metering-operator.v${METERING_OPERATOR_IMAGE_TAG}
EOF

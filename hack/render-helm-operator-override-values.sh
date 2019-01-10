#!/bin/bash

set -e
set -u

cat <<EOF
image:
  repository: ${METERING_OPERATOR_IMAGE}
  tag: ${METERING_OPERATOR_IMAGE_TAG}
EOF

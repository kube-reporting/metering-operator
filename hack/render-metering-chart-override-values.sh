#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
reporting-operator:
  spec:
    image:
      tag: $1
EOF

#!/bin/bash
set -e

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${1:?Must be the namespace which pod\'s logs you want to capture}"

capture_pod_logs "$1"

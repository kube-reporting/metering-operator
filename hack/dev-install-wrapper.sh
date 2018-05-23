#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..

if [ $# -lt 2 ]; then
    echo "Usage: $0 [metering-helm-operator-image-tag] [install-script] [args]"
    echo "Must specify at least two args, metering-helm-operator image tag, and the install script to run"
    exit 1
***REMOVED***

HELM_OPERATOR_IMAGE_TAG="$1"
shift

TMP="$(mktemp -d)"

trap "rm -rf $TMP" EXIT SIGINT

"$ROOT_DIR/hack/create-metering-manifests.sh" "$HELM_OPERATOR_IMAGE_TAG" "$TMP"

"$@"

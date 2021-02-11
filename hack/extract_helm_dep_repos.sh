#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

"$FAQ_BIN" -f yaml -o json -M -c -r -p=false \
        '.dependencies[].repository | ltrimstr("file://")' \
        "$@" \
        | grep '\S'

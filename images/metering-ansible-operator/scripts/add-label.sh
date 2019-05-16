#!/bin/bash

set -e
set -u

: "${FAQ_BIN:=faq}"

export PRUNE_LABEL_KEY=${1:?}
export PRUNE_LABEL_VALUE=${2:?}

# shellcheck disable=SC2016
"$FAQ_BIN" \
    -f yaml -o yaml -r -M \
    'select(. != null) | .metadata.labels[$ENV.PRUNE_LABEL_KEY]=$ENV.PRUNE_LABEL_VALUE'

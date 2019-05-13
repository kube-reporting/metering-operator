#!/bin/bash

set -e
set -u

: "${FAQ_BIN:=faq}"

export OWNER_API_VERSION=${1:?}
export OWNER_KIND=${2:?}
export OWNER_NAME=${3:?}
export OWNER_UID=${4:?}
export BLOCKER_OWNER_DELETION=${5:?}

# shellcheck disable=SC2016
"$FAQ_BIN" \
    -f yaml -o yaml -r -M \
    'select(. != null) | .metadata.ownerReferences |= [{apiVersion: $ENV.OWNER_API_VERSION, kind: $ENV.OWNER_KIND, uid: $ENV.OWNER_UID, name: $ENV.OWNER_NAME, controller: true, blockOwnerDeletion: ($ENV.BLOCKER_OWNER_DELETION == "true")}]'

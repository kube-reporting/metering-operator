#!/bin/bash

# This script is called by the Make***REMOVED***le when determining if the generated
# client code needs to be regenerated.
# We use the output of this script as the Make***REMOVED***le target which runs
# ./hack/update-codegen.sh, since we don't necessarily know all the outputted
# ***REMOVED***les ahead of time (or rather, it's pain to hardcode the ***REMOVED***les).
# If the client has been generated before, this will output the list of ***REMOVED***les
# generated which are used as the target, otherwise it will print nothing,
# and you will need to run ./hack/update-codegen.sh manually.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$SCRIPT_DIR/.."
DIRS=("${ROOT_DIR}/pkg/generated")

for DIR in "${DIRS[@]}"; do
    OUTPUT=$(***REMOVED***nd "$(realpath ${DIR})" -type f -name '*.go')
    echo "$OUTPUT"
done

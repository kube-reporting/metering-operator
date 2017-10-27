#!/bin/bash

# This script is called by the Makefile when determining if the generated
# client code needs to be regenerated.
# We use the output of this script as the Makefile target which runs
# ./hack/update-codegen.sh, since we don't necessarily know all the outputted
# files ahead of time (or rather, it's pain to hardcode the files).
# If the client has been generated before, this will output the list of files
# generated which are used as the target, otherwise it will print nothing,
# and you will need to run ./hack/update-codegen.sh manually.

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GENERATED_DIR=$(realpath "${DIR}/../pkg/generated")
OUTPUT=$(find "$GENERATED_DIR" -type f -name '*.go')

echo "$OUTPUT"


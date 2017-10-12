#!/bin/bash

# This script is called by the Makefile when determining if the generated
# client code needs to be regenerated.
# We use the output of this script as the Makefile target which runs
# ./hack/update-codegen.sh, since we don't necessarily know all the outputted
# files ahead of time (or rather, it's pain to hardcode the files).
# If the client has been generated before, this will output the list of files
# generated which are used as the target, otherwise it prints the date in unix
# time which will never exist as an output file, resulting in a rebuild

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GENERATED_DIR=$(realpath "${DIR}/../pkg/generated")
OUTPUT=$(find "$GENERATED_DIR" -type f -name '*.go')

# If there's no generated files in the output, then touch all the source files
# to force Make to re-build the generated files target.
if [ -z "$OUTPUT" ]; then
    date +'%s'
else
    echo "$OUTPUT"
fi


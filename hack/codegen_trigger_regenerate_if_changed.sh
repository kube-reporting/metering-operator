#!/bin/bash

# This script is called by the Make***REMOVED***le when determining if the generated
# client code needs to be regenerated.
# We use the output of this script as the Make***REMOVED***le target which runs
# ./hack/update-codegen.sh, since we don't necessarily know all the outputted
# ***REMOVED***les ahead of time (or rather, it's pain to hardcode the ***REMOVED***les).
# If the client has been generated before, this will output the list of ***REMOVED***les
# generated which are used as the target, otherwise it prints the date in unix
# time which will never exist as an output ***REMOVED***le, resulting in a rebuild

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GENERATED_DIR=$(realpath "${DIR}/../pkg/generated")
OUTPUT=$(***REMOVED***nd "$GENERATED_DIR" -type f -name '*.go')

# If there's no generated ***REMOVED***les in the output, then touch all the source ***REMOVED***les
# to force Make to re-build the generated ***REMOVED***les target.
if [ -z "$OUTPUT" ]; then
    date +'%s'
***REMOVED***
    echo "$OUTPUT"
***REMOVED***


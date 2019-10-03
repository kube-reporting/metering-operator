#!/bin/bash

set -e

SOURCE_IMAGE=$1
OUTPUT_IMAGE=$2

if [[ -z "$SOURCE_IMAGE" ]]; then
        echo "must pass a source image as the first arg"
        exit 1
fi

if [[ -z "$OUTPUT_IMAGE" ]]; then
        echo "must pass the output image as the second arg"
        exit 1
fi

DOCKER_COMMAND=${DOCKER_COMMAND:-docker}

set -x
"$DOCKER_COMMAND" pull "$SOURCE_IMAGE"
"$DOCKER_COMMAND" tag "$SOURCE_IMAGE" "$OUTPUT_IMAGE"
"$DOCKER_COMMAND" push "$OUTPUT_IMAGE"

set +x
echo "Mirrored $SOURCE_IMAGE to $OUTPUT_IMAGE"

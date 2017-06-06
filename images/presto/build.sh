#!/bin/bash
# Builds image for presto
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/fest-data-demo/presto:0.1"}

docker build -t "${IMAGE_NAME}" ${DIR}

#!/bin/bash
# Builds image for hadoop base image
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/fest-data-demo/hadoop:0.1"}

docker build -t "${IMAGE_NAME}" ${DIR}

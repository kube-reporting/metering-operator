#!/bin/bash
# Builds image for promsum
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/coreos/promsum:0.2"}

docker build -t "${IMAGE_NAME}" ${DIR}

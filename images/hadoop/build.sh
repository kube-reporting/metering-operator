#!/bin/bash
# Builds image for hadoop base image
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/coreos/chargeback-hadoop:0.2"}

docker build -t "${IMAGE_NAME}" ${DIR}

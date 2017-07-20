#!/bin/bash
# Builds image for promsum
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/coreos/chargeback:0.1"}

docker build -t "${IMAGE_NAME}" ${DIR}

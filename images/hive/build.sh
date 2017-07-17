#!/bin/bash
# Builds image for hive
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME=${@:-"quay.io/fest-data-demo/hive:0.2"}

docker build -t "${IMAGE_NAME}" ${DIR}

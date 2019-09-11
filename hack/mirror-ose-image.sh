#!/bin/bash

set -e

IMAGE_NAME=$1
CLUSTER_REGISTRY_URL=$2

OSE_IMAGE_REPO="${OSE_IMAGE_REPO:-"brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888"}"

if [[ -z $IMAGE_NAME ]]; then
        echo "must pass a image name as the first arg"
        exit 1
fi

if [[ -z $CLUSTER_REGISTRY_URL ]]; then
        echo "must pass the cluster registry hostname as the second arg"
        exit 1
fi

DOCKER_COMMAND=${DOCKER_COMMAND:-docker}

set -x
"$DOCKER_COMMAND" pull "$OSE_IMAGE_REPO/$IMAGE_NAME"
"$DOCKER_COMMAND" tag "$OSE_IMAGE_REPO/$IMAGE_NAME" "$CLUSTER_REGISTRY_URL/$IMAGE_NAME"
"$DOCKER_COMMAND" push "$CLUSTER_REGISTRY_URL/$IMAGE_NAME"

set +x
echo "$IMAGE_NAME is added to the cluster"

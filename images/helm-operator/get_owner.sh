#!/bin/bash

set -e

: ${MY_POD_NAME:?}
: ${MY_POD_NAMESPACE:?}

kubectl -n "$MY_POD_NAMESPACE" get pod "$MY_POD_NAME" -o json > /tmp/my_pod.json
export MY_RS_NAME=$(jq '.metadata.ownerReferences[].name' -r /tmp/my_pod.json)

kubectl -n "$MY_POD_NAMESPACE" get rs "$MY_RS_NAME" -o json > /tmp/my_rs.json
export MY_DEPLOYMENT_NAME=$(jq '.metadata.ownerReferences[].name' -r /tmp/my_rs.json)

kubectl -n "$MY_POD_NAMESPACE" get deployment "$MY_DEPLOYMENT_NAME" -o json > /tmp/my_deployment.json
export MY_DEPLOYMENT_UID=$(jq '.metadata.uid' -r /tmp/my_deployment.json)


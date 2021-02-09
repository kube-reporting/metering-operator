#!/bin/bash

# Note: before running, ensure that the correct registry service is exposed
# The following would expose the redhat-operators CatalogSource service:
# kubectl -n openshift-marketplace port-forward svc/redhat-operators 50051

CHANNEL="${CHANNEL:-"4.8"}"

read -r -d '' JQ_PROGRAM << EOM
 .spec.install.spec.deployments[0].spec.template.spec.containers[]
 | select(.name=="operator")
 | .env
 | map(select(.name | match(".*_IMAGE")))
 # | map(.value |= ltrimstr("image-registry.openshift-image-registry.svc:5000/"))
 # | map(.value |= (split(":"))[1])
 # |  map(.name |= gsub("_IMAGE"; "_IMAGE_TAG"))
 | .[]
 | "export " + .name + "=" + .value
EOM

grpcurl -plaintext -d '{"pkgName": "metering-ocp", "channelName": "'"$CHANNEL"'"}' localhost:50051 api.Registry/GetBundleForChannel  \
    | jq '.csvJson' -r \
    | jq "$JQ_PROGRAM" -r


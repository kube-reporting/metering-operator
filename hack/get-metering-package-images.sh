#!/bin/bash

# kubectl -n openshift-marketplace port-forward svc/art-applications 50051

grpcurl -plaintext -d '{"pkgName": "metering-ocp", "channelName": "4.2"}' localhost:50051 api.Registry/GetBundleForChannel  | jq '.csvJson' -r | jq '.spec.install.spec.deployments[0].spec.template.spec.containers[] | select(.name=="operator") | .env | map(select(.name | match(".*_IMAGE"))) | map(.value |= (. | ltrimstr("image-registry.openshift-image-registry.svc:5000/"))) | .[] | "export " + .name + "=" + .value' -r

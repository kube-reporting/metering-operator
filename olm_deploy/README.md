# Overview

This directory contains resources related to deploying Metering through the Operator Lifecycle Manager (OLM) using the most up-to-date operators bundle.

## Overview of the registry Dockerfile(s)

Before we can deploy Metering through OLM, we need to create a `CatalogSource` custom resource, which is a way to expose a collection of an operator's packages and channels. The goal here is create a `CatalogSource` that references the newest version of the metering-ansible-operator's metadata.

We have two Dockerfiles that achieve this goal:

- Dockerfile.registry: Creates a registry image, which contains the latest metering-ansible-operator manifest bundle, but adds additional scripts that manipulate the CSV manifests. This image cannot be run as a stand-alone image.
- Dockerfile.registry.dev: Creates a registry image containing the repository's manifest bundle, loads that bundle in a sqlite database, and exposes that database using a gRPC interface. This is useful for when you want to test Metering with the default image listed in the CSV.

### Building the Dockerfile.registry image

**Note**: The following script defaults to using podman as the container runtime.

To override the default container runtime, set the following environment variable:

```bash
export CONTAINER_RUNTIME="..."
```

Run the following command, providing a reference to a repository and tag as the first argument:

```bash
./hack/push-olm-manifests.sh quay.io/<namespace>/<repository>:<tag>
```

## Overview of manifests

Inside of this directory is a set of registry-related manifests:

- manifests/deployment.yaml: Base template of a deployment responsible for manipulating the CSV images in an initContainer and serving those manipulated CSV images in a sqlite database.
- manifests/service.yaml: Base template of a service that exposes the gRPC interface for the sqlite database.

By default, the images listed in the manifest bundle CSV default to the registry.svc.ci.openshift.org registry. In the case any of the following environment variables have been specified, the initContainer in the deployment manifest fill substitute that image with the value stored in the environment variable.

Here is the list of configurable environment variables:

- $IMAGE_FORMAT: variable exposed by the CI environment that tags all of the metering-operator + its operand images into a shared image stream.
- $IMAGE_METERING_ANSIBLE_OPERATOR: container image for the metering-ansible-operator.
- $IMAGE_METERING_REPORTING_OPERATOR: container image for the reporting-operator.
- $IMAGE_METERING_PRESTO: container image for presto.
- $IMAGE_METERING_HIVE: container image for hive.
- $IMAGE_METERING_HADOOP: container image for hadoop.
- $IMAGE_GHOSTUNNEL: container image for ghostunnel.
- $IMAGE_OAUTH_PROXY: container image for Openshift oauth-proxy.
- $IMAGE_METERING_ANSIBLE_OPERATOR_REGISTRY: container image for the metering-ansible-operator manifest bundle.

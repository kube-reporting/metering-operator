# Testing Metering with OCP images

This document covers how to mirror OCP images from the internal Red Hat container image repository to your in-cluster registry, and how to install Metering so that it uses the OCP images.

# Setup

Before we install, you need to correctly configure your docker daemon, and mirror the images.

# Configure access to your cluster's in-cluster image registry

Next, you need to login to your cluster's in-cluster image registry.
Because the in-cluster cluster registry is using a self-signed certificate, you will need to configure your docker daemon to trust the cluster CA:

Get your registry hostname by running the following:

```
oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}'
```

And then follow the instructions in https://docs.docker.com/registry/insecure/ to add the registry to your insecure registry list.
You can also copy the cluster's CA into the correct directory within `/etc/docker/certs.d/` by following the instructions in the same page.

# Mirror images from Brew into your cluster's in-cluster image registry

Make sure you're connected to the Red Hat in-cluster network (VPNed or otherwise) and run the `hack/mirror-ose-images-into-cluster.sh` script in the repo.
This script does a few things:

- Sets up a namespace for pushing the images into.
- Creates a serviceaccount and grants it permissions to push images to the in-cluster registry.
- Uses the serviceaccount to `docker login` to the in-cluster registry.
- Pulls images from the Red Hat internal registry to your local workstation.
- Pushes them into your in-cluster image registry.

```
./hack/mirror-ose-images-into-cluster.sh
```

# Install

Set your `$METERING_CR_FILE` variable as documented in the [manual installation guide][manual-install] and then do a manual install using the following command:

```
INSTALLER_MANIFESTS_DIR=manifests/deploy/ocp-testing/metering-ansible-operator ./hack/openshift-install.sh
```

[manual-install]: manual-install.md

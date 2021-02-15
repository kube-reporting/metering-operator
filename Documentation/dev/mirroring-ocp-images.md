# Mirroring OCP images into your cluster

This guide will walk through:

- Configuring your OCP 4 cluster's registry to be accessible from outside the cluster
- Configuring docker so that you can pull images from the Red Hat internal docker registry: `brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888`
- Running the mirroring script provided to pull images from brew-pulp and push them into your OCP 4 cluster's in-cluster registry.

## Configure access to your cluster's in-cluster image registry

First, you need to ensure your registry is accessible outside the cluster so we can push images into it.
Follow <https://docs.openshift.com/container-platform/4.1/registry/securing-exposing-registry.html> to do this.

The short version is to run the following command:

```bash
oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge
```

Your registry is always secured with authentication, so exposing it outside the cluster should be safe.

Next, you need configure docker to trust your registry.
Because the in-cluster cluster registry is using a self-signed certificate, you will need to configure your docker daemon to trust the cluster CA:

Get your registry hostname by running the following:

```bash
oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}'
```

And then follow the instructions in <https://docs.docker.com/registry/insecure/> to add the registry to your insecure registry list.
You can also copy the cluster's CA into the correct directory within `/etc/docker/certs.d/` by following the instructions in the same page.

## Mirror images from Brew into your cluster's in-cluster image registry

Make sure you're connected to the Red Hat in-cluster network (VPNed or using an ethernet connection on the Red Hat network) and run the [`hack/mirror-ose-images-into-cluster.sh`](../../hack/mirror-ose-images-into-cluster.sh) script in the repo.
This script does a few things:

- Sets up a namespace for pushing the images into.
- Creates a serviceaccount and grants it permissions to push images to the in-cluster registry.
- Uses the serviceaccount to `docker login` to the in-cluster registry.
- Pulls images from the Red Hat internal registry to your local workstation.
- Pushes them into your in-cluster image registry.

```bash
./hack/mirror-ose-images-into-cluster.sh
```

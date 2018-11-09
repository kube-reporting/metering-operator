# Building OCP Images

Building OCP images requires being on the Red Hat network so you can access our internal image and rpm repositories.
Additionally, because the images are generally built using automation thats already setup correctly for these builds, we need to use some additional tools to make building images based on RHEL succeed.

Pre-reqs:
- All the regular developer tools
- Docker
- [imagebuilder](https://github.com/openshift/imagebuilder)
- On the Red Hat network (VPNed or otherwise)

## Pull base images and rename them

Run the following script to setup the base images we need locally:

```
hack/ocp-util/ocp-image-pull-and-rename.sh
```

## Build OCP images using imagebuilder

Passing `OCP_BUILD=true` to our make invocation will tell it it to build our images using the `.rhel` Docker***REMOVED***les and updates the image names to match the registry.access.redhat.com names.
Passing `USE_IMAGEBUILDER=true` tells it to use `imagebuilder` instead of `docker` as a Docker client, allowing us to override the RPM repos and subscription-manager con***REMOVED***guration inside the image build to allow installing packages without needing to build on a RHEL machine by using yum composes that the regular OCP release pipeline does.
Passing `RELEASE_TAG=v4.0` overrides using the version in the `VERSION` ***REMOVED***le so that we can tag our images using the OCP release version.

```
make docker-build-all OCP_BUILD=true USE_IMAGEBUILDER=true RELEASE_TAG=v4.0
```

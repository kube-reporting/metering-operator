# Developer Guide

This document describes setting up your environment, as well as installing Metering.

## Development Dependencies

- Go 1.13 or higher
- Helm CLI up to 2.8.1
- Make
- Docker
- jq
- [faq](https://github.com/jzelinskie/faq) 0.0.5 or newer
  - For Fedora 29, Rawhide, and CentOS 7 you can use the following copr repo: https://copr.fedorainfracloud.org/coprs/ecnahc515/faq/
  - For mac: `brew tap jzelinskie/faq && brew install faq`
  - Or you can download the release binaries directly from Github: https://github.com/jzelinskie/faq/releases
- realpath

If you're using MacOS with homebrew you can install all of these using the following:

```bash
$ brew tap jzelinskie/faq
$ brew install go make docker coreutils jq faq python@3
[ and install helm 2.8.1 in your path from https://kubernetes-helm.storage.googleapis.com/helm-v2.8.1-darwin-amd64.tar.gz ]
```

## Setup

Clone this repository:

```bash
git clone https://github.com/kube-reporting/metering-operator
```

## Building

To build the reporting-operator binary using your local Go:

```bash
make reporting-operator-bin
```

If you want to build docker images locally:

```bash
make docker-build-all
```

If you want to build OCP images locally see [the Building OCP images document](ocp-images.md).

## Running tests and checks locally

To run unit tests:

```bash
make unit
```

To run the validation steps CI performs:

```bash
make verify
```

To verify the vendored dependencies are in order:

```bash
make vendor
```

### Running the e2e tests locally

#### Overview

There are two top-level tests that aim towards testing various Metering configurations against the newest version, and testing the upgrability from a previous version.

At a high-level view, the e2e suite is composed of two objects:

- The `deployframework` object, which is the entrypoint to the testing suite, provides all the initialized clientsets, creates any pre-installation resources, and more.
- The `deployerCtx` object, which is responsible for the state and management of a single Metering installation.

Before we deploy any Metering instances, we first need to create the `CatalogSource` custom resource, which is a way to expose a collection of an operator's packages and channels, that points to the newest version of the metering-ansible-operator's metadata.

We do this by building a registry image, which contains the latest metering-ansible-operator manifest bundle, scripts to manipulate the CSV manifest and a few of the operator-registry binaries that we use throughout the e2e suite.

For more information on the operator-registry binaries we utilize, check out the [operator-registry overview](https://github.com/operator-framework/operator-registry#overview).

#### Building the manifest bundle registry image

In the testing suite, we deploy all Metering instances using OLM, which requires building and pushing the latest version of the metering-ansible-operator's manifest bundle.

Define your CONTAINER_RUNTIME, for instance:
```bash
export CONTAINER_RUNTIME=docker
```

Run the following command, providing a reference to a repository and tag as the first argument:

```bash
./hack/push-olm-manifests.sh quay.io/tflannag/metering-registry:latest
```

This will build the `./olm_deploy/Dockerfile.registry` registry image, which copies over the OLM-related manifest bundle (e.g. `manifests/deploy/openshift/olm/bundle`) for further processing later down the line.

Once that image has been pushed to an image registry, like quay or docker.io, set the following environment variable to point to that newly build image:

```bash
export METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY="quay.io/tflannag/metering-registry:latest"
```

**Note**: In the case where the `$IMAGE_FORMAT` environment variable is set (i.e. in a CI environment), the values that we pull out of that variable will override the `$METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY` value.

#### Running the e2e suite

There are a couple of different ways to run the tests in the e2e suite locally.

- `make e2e`: Runs the e2e suite without any altered workflows.
- `make e2e-dev`: Runs the e2e suite, but doesn't teardown or clean up any Metering installations.
- `make e2e-local`: Builds and runs the metering-ansible-operator and reporting-operator images as containers locally.
- `make e2e-local-dev`: Builds and runs the metering-ansible-operator and reporting-operator images as containers locally, but doesn't teardown or clean up any Metering installations.

Before running any of those Makefile targets, ensure the following variables are exported:

**Note**: only the first environment variable, which controls the manifest bundle registry image, is required.

- `$METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY`
- `$METERING_OPERATOR_IMAGE_REPO`
- `$METERING_OPERATOR_IMAGE_TAG`
- `$REPORTING_OPERATOR_IMAGE_REPO`
- `$REPORTING_OPERATOR_IMAGE_TAG`

When interacting with these Makefile targets, we expose several variables to help customize and provide additional flexibility. Here are general ones you may to specify:

- `$TEST_OUTPUT_PATH`: Controls where all of the testing artifacts are stored. Defaults to a /tmp/ directory that gets created.
- `$TEST_LOG_LEVEL`: Controls the log verbosity that gets logged to files and stdout. Defaults to "debug".
- `$METERING_OLM_SUBSCRIPTION_CHANNEL`: Specifies what channel of Metering should be deployed when creating Subscription custom resources. Defaults to the "4.6" channel.
- `$EXTRA_TEST_FLAGS`: Specifies any additional `go test` flags that should be run. Useful for when you want to only run the manual metering tests, or just the upgrade ones.

For a full list, check out the [./hack/e2e.sh bash script](https://github.com/kube-reporting/metering-operator/blob/master/hack/e2e.sh).

##### Examples

Overriding the default test artifacts directory:

```bash
make e2e TEST_OUTPUT_PATH=$PWD/metering_test_output
```

Overriding the default subscription channel:

```bash
make e2e METERING_OLM_SUBSCRIPTION_CHANNEL="latest"
```

Providing extra test flags:

```bash
make e2e EXTRA_TEST_FLAGS="-race -run TestManualMeteringInstall"
```

## Go Dependencies

We use Go modules for managing dependencies.

`go mod` installs dependencies into the `vendor/` directory at the
root of the repository, and to ensure everyone is using the same dependencies,
and ensure that if dependencies disappear, we commit the contents of `vendor/`
into git.

### Adding new dependencies

To add a new dependency, you can do the following:

```bash
go get <dependency_repo_url>@<version_of_dependency>
```

You can learn more about version specification here: [using-go-modules](https://blog.golang.org/using-go-modules).

`go get` will modify `go.mod` and `go.sum` for you.

Run `make vendor` after adding dependencies with `go get` and before committing.

When committing new dependencies, please use the following guidelines:

- Always commit changes to dependencies separately from other changes.
- Use one commit for changes to `go.mod`, and another commit for changes to
  `go.sum` and `vendor/`.' Commit messages should be in the following forms:
  - `go.mod: Add new dependency $your_new_dependency`
  - `go.sum,vendor: Add new dependency $your_new_dependency`

## Helm templates

If you have added a new Helm Chart and would like to render the template to check the values and nesting within the yaml file you can run:

```bash
helm template CHART_DIR -x PATH_TO_TEMPLATE/file.yaml
```

Additionally, you can use the `--set` flag to assign a variable used throughout a helm template with a specific value to further test the chart's templating.

## Developer install

Developers should generally use the [manual-install guide](../manual-install.md) as it offers the most flexibility when installing.

If you need a minimal storage configuration with no external dependencies, use the [manifests/metering-config/hdfs-minimal.yaml](../../manifests/metering-config/hdfs-minimal.yaml) example configuration.

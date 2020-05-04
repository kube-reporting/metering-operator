# Developer Guide

This document describes setting up your environment, as well as installing Metering.

## Development Dependencies

- Go 1.13 or higher
- Helm CLI 2.6.2 or higher minor version
- Make
- Docker
- jq
- [faq](https://github.com/jzelinskie/faq) 0.0.5 or newer
  - For Fedora 29, Rawhide, and CentOS 7 you can use the following copr repo: https://copr.fedorainfracloud.org/coprs/ecnahc515/faq/
  - For mac: `brew tap jzelinskie/faq && brew install faq`
  - Or you can download the release binaries directly from Github: https://github.com/jzelinskie/faq/releases
- realpath
- [operator-courier](https://github.com/operator-framework/operator-courier)
  - `pip3 install operator-courier`

If you're using MacOS with homebrew you can install all of these using the
following:

```bash
$ brew tap jzelinskie/faq
$ brew install go kubernetes-helm make docker coreutils jq faq python@3
$ pip3 install operator-courier
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

## Running local tests

To run unit tests:

```bash
make unit
```

To run the validation steps CI does:

```bash
make verify
```

### Running e2e/integration tests locally

There's 2 ways to run integration and e2e tests locally.
The first option runs the main operators locally (and the rest of the components into the cluster), and runs tests against the local reporting-operator, and the second option deploys everything into the cluster, and runs tests against the reporting-operator in the cluster.

The first option can be broken down in a few steps:

- Build the metering-operator docker image
- Build the reporting-operator binary locally
- Run a customized manual install for tests, skipping metering-operator and reporting-operator
- Run metering-operator locally via Docker
- Run reporting-operator locally as a native Go binary
- Configure everything to properly communicate
- Run tests against the local reporting-operator

You can run one of the following commands to run either e2e or integration tests locally which will do the above steps, testing against a local reporting-operator:

```bash
make e2e-local TEST_OUTPUT_PATH=/tmp/metering_e2e_output
make integration-local TEST_OUTPUT_PATH=/tmp/metering_integration_output
```

The second option is similar, but doesn't run reporting-operator or metering-operator locally, but instead deploys them into the cluster just like CI does.
The steps can be broken down into:

- Run a customized manual install for tests
- Run tests against the deployed reporting-operator running in the cluster

Replace `pr-1234` with your image tag (usually built by CI), and run one of the following commands to run e2e or integration tests locally against fully deployed metering stack:

```bash
make e2e REPORTING_OPERATOR_IMAGE_TAG=pr-1234 METERING_OPERATOR_IMAGE_TAG=pr-1234 TEST_OUTPUT_PATH=/tmp/metering_e2e_output
make integration REPORTING_OPERATOR_IMAGE_TAG=pr-1234 METERING_OPERATOR_IMAGE_TAG=pr-1234 TEST_OUTPUT_PATH=/tmp/metering_integration_output
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

## Developer install

Developers should generally use the [manual-install guide](../manual-install.md) as it offers the most flexibility when installing.
If you need a minimal storage configuration with no external dependencies, use the [manifests/metering-config/hdfs-minimal.yaml](../../manifests/metering-config/hdfs-minimal.yaml) example configuration.


# Overview

The `deploy-metering` utility aims to emulate the install/uninstall logic that's currently done in the `hack/install.sh` and `hack/uninstall.sh` shell scripts.

When interacting with the `deploy-metering` binary, you have the choice of using the command flags available, or their respective environment variable.

In the case where both the flag, and the environment variable are specified, the [flag value takes precedence](#example-of-precedence) over the value stored in the environment variable.

## Building the binary

```bash
make bin/deploy-metering
```

### Prerequisites

With the addition of the deploy-metering CLI, we can assign a default (typically an empty value) value for most of the configuration options.

This means the only required flags/environment variables that needs to be set, are the namespace in which the Metering stack will be managed and the path to the `MeteringConfig` custom resource file.

Before deploying metering, ensure that either of the following options are set:

```bash
export METERING_NAMESPACE="metering"
export METERING_CR_FILE=<path to custom MeteringConfig custom resource>
```

Alternatively, run the CLI using the corresponding flags:

```bash
./bin/deploy-metering install --namespace metering --meteringconfig <path to custom MeteringConfig custom resource>
```

It's important to note that when the path to the `MeteringConfig` custom resource is unspecified, the package uses the default `MeteringConfig` manifest resource, which contains an invalid storage specification.

If you do not set that environment variable or flag, the installation will succeed, but it will fail the validation check in the metering-ansible-operator Ansible role due to an invalid user-provided storage configuration.

### General Configuration Options

- `$METERING_NAMESPACE`/`--namespace`: the namespace to deploy metering resources on.
- `$METERING_CR_FILE`/`--metering-cr`: the relative (or absolute) path to a customized `MeteringConfig` custom resource.
- `$DEPLOY_MANIFESTS_DIR`/`--deploy-manifests-dir`: the relative (or absolute) path to the `manifests/deploy/` directory. Set this when working outside of the metering-operator cloned repository, or when the `bin/deploy-metering` binary has been moved to the user's `/usr/local/bin` directory.
- `$DEPLOY_PLATFORM`/`--deploy-platform`: the platform in which metering is deploy on. The supported values include "upstream" and "openshift".
- `$METERING_DEPLOY_LOG_LEVEL`/`--log-level`: controls the verbosity of the logs.

#### Install-Only Configuration Options

- `$SKIP_METERING_OPERATOR_DEPLOYMENT`/`--skip-metering-operator-deployment`: create only the `$METERING_NAMESPACE` namespace, the metering CRDs, and the `MeteringConfig` CR.
- `$METERING_OPERATOR_IMAGE_REPO`/`--repo`: override the image repository used in the metering-ansible-operator container images.
- `$METERING_OPERATOR_IMAGE_TAG`/`--tag`: override the image tag used in the metering-ansible-operator container images.

**Note**: You need to set both the image repository and tag to override any of the container images.

#### Uninstall-Only Configuration Options

- `$METERING_DELETE_NAMESPACE`/`--delete-namespace`: defaults to false.
- `$METERING_DELETE_CRDS`/`--delete-crd`: defaults to false.
- `$METERING_DELETE_CRB`/`--delete-crb`: defaults to false.
- `$METERING_DELETE_PVCS`/`--delete-pvc`: defaults to true.
- `$METERING_DELETE_ALL`/`--delete-all`: defaults to false. This sets all the internal states that rely on the value of the environment variables listed above, to true.

### Example of precedence

```bash
$ export METERING_NAMESPACE="metering"
$ ./bin/deploy-metering install --namespace test
INFO[09-17-2019 14:54:42] Setting the log level to info                 app=deploy
INFO[09-17-2019 14:54:42] Metering Deploy Namespace: test               app=deploy
...
```

### Install Usage Examples

**Note**: You can either use the flags, or the environment variables when running any of the following examples. Also, specify a customized `MeteringConfig` custom resource before running any of these examples.

#### Vanilla install on Openshift/OCP

```bash
export METERING_NAMESPACE="metering"
./bin/deploy-metering install
```

#### Vanilla install on upstream

```bash
export METERING_NAMESPACE="metering"
export DEPLOY_PLATFORM="upstream"
./bin/deploy-metering install
```

#### Skip installing the metering operator

```bash
export METERING_NAMESPACE="metering"
export SKIP_METERING_OPERATOR_DEPLOYMENT="true"
./bin/deploy-metering install
```

#### Override the default images

```bash
export METERING_NAMESPACE="metering"
export METERING_OPERATOR_IMAGE_REPO="<replace with image repository>"
export METERING_OPERATOR_IMAGE_TAG="<replace with image tag>"
./bin/deploy-metering install
```

#### Specify an explicit manifest location

```bash
$ pwd
/home/tflannag/go/src/github.com/kube-reporting/metering-operator
$ cd ../
$ pwd
/home/tflannag/go/src/github.com/kube-reporting
$ export METERING_NAMESPACE="metering"
$ export DEPLOY_MANIFESTS_DIR="metering-operator/manifests/deploy"
$ ./metering-operator/bin/deploy-metering install
```

### Uninstall Usage Examples

**Note**: set the `$METERING_CR_FILE` before running any of these examples.

#### Vanilla uninstall on Openshift/OCP

```bash
export METERING_NAMESPACE="metering"
./bin/deploy-metering uninstall
```

#### Vanilla uninstall for upstream

```bash
export METERING_NAMESPACE="metering"
export DEPLOY_PLATFORM="upstream"
./bin/deploy-metering uninstall
```

#### Delete cluster role/role bindings

```bash
export METERING_NAMESPACE="metering"
export METERING_DELETE_CRB="true"
./bin/deploy-metering uninstall
```

**Note**: only the `$METERING_NAMESPACE` and the metering CRDs should be skipped.

#### Delete all the Metering cluster roles/role bindings and namespace

```bash
export METERING_NAMESPACE="metering"
export METERING_DELETE_NAMESPACE="true"
export METERING_DELETE_CRB="true"
./bin/deploy-metering uninstall
```

#### Delete all traces of the metering stack during an uninstall

```bash
export METERING_NAMESPACE="metering"
export METERING_DELETE_ALL="true"
./bin/deploy-metering uninstall
```

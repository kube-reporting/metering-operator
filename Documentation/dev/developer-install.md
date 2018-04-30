# Dev Installation

If you want to install metering without ALM, using what's currently in master, first start by cloning the repo.
Next, decide which namespace you want to install Metering into, and set the `METERING_NAMESPACE` environment variable to the namespace you want to use.
By default, if it's unset, it will use the `metering` namespace.

Run the following commands:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/install.sh
```

To uninstall simply:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/uninstall.sh
```

## Customize installation

If you wish to customize the installation, such as to modify configuration
options, change the image tag or repository, then you can use a custom
`metering` resource. To start, copy the default metering resource to a
separate file that we can modify:

```
$ cp manifests/metering-config/default.yaml metering-custom.yaml
```

For developers, the most common change is modifying the image tag, config, and resource limits.
Take a look at `manifests/metering-config/custom-values.yaml` to get an
idea of what you can modify that relates to configuration and resource limits, and
`manifests/metering-config/latest-versions.yaml` to see  how to change the
image tag of each component.

```
$ export METERING_NAMESPACE=metering-$USER
$ export METERING_CR_FILE=metering-custom.yaml
$ ./hack/install.sh
```

### Using images built by Jenkins

If you have a PR or branch being built my Jenkins, you can use the images it's publishing from each build to test out the changes that aren't in master yet.
For details on the image tag format, please follow the instructions in our [jenkins guide](jenkins.md#using-images-built-by-jenkins).


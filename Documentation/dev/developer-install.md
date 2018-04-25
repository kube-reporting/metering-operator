# Dev Installation

If you want to install chargeback without ALM, using what's currently in master, first start by cloning the repo.
Next, decide which namespace you want to install Chargeback into, and set the `CHARGEBACK_NAMESPACE` environment variable to the namespace you want to use.
By default, if it's unset, it will use the `chargeback` namespace.

Run the following commands:

```
$ export CHARGEBACK_NAMESPACE=chargeback-$USER
$ ./hack/install.sh
```

To uninstall simply:

```
$ export CHARGEBACK_NAMESPACE=chargeback-$USER
$ ./hack/uninstall.sh
```

## Customize installation

If you wish to customize the installation, such as to modify configuration
options, change the image tag or repository, then you can use a custom
`chargeback` resource. To start, copy the default chargeback resource to a
separate file that we can modify:

```
$ cp manifests/chargeback-config/default.yaml chargeback-custom.yaml
```

For developers, the most common change is modifying the image tag, config, and resource limits.
Take a look at `manifests/chargeback-config/custom-values.yaml` to get an
idea of what you can modify that relates to configuration and resource limits, and
`manifests/chargeback-config/latest-versions.yaml` to see  how to change the
image tag of each component.

```
$ export CHARGEBACK_NAMESPACE=chargeback-$USER
$ export CHARGEBACK_CR_FILE=chargeback-custom.yaml
$ ./hack/install.sh
```

### Using images built by Jenkins

If you have a PR or branch being built my Jenkins, you can use the images it's publishing from each build to test out the changes that aren't in master yet.
For details on the image tag format, please follow the instructions in our [jenkins guide](jenkins.md#using-images-built-by-jenkins).


# Dev Installation

If you want to install chargeback without ALM, using what's currently in master, first clone the repo and then run the following commands:

```
$ export CHARGEBACK_NAMESPACE=chargeback
$ ./hack/install.sh
```

To uninstall simply:

```
$ export CHARGEBACK_NAMESPACE=chargeback
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
$ export CHARGEBACK_NAMESPACE=chargeback
$ export CHARGEBACK_CR_FILE=chargeback-custom.yaml
$ ./hack/install.sh
```


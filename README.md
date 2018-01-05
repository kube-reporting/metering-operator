# Chargeback

Chargeback records historical cluster usage, and can generate usage reports showing usage breakdowns by pod or namespace over arbitrary time periods.

## Getting Started

For installation instructions, please see the [installation](Documentation/installation.md) portion of the documentation.

### Dev installation

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

See the above installation guide for links to configuring and using Chargeback.

### Development Dependencies

- Go 1.8 or higher
- Helm 2.6.2
- Make
- Docker
- Glide


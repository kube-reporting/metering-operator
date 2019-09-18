# Testing Metering with OCP images

This document covers how to mirror OCP images from the internal Red Hat container image repository to your in-cluster registry, and how to install Metering so that it uses the OCP images.

# Mirror OCP images into your cluster

Follow the [Mirroring OCP images into your cluster](mirroring-ocp-images.md) guide for instructions.

# Install

Set your `$METERING_CR_FILE` variable as documented in the [manual installation guide][manual-install] and then do a manual install using the following command:

```
INSTALLER_MANIFESTS_DIR=manifests/deploy/ocp-testing/metering-ansible-operator ./hack/openshift-install.sh
```

[manual-install]: manual-install.md

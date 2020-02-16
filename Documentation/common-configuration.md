# Common configuration

This document contains example configurations for configuration that spans one or more components.

## Storage

Storage has it's own documentation page at [configuring storage](configuring-storage.md).

## Resource requests and limits

You can adjust the cpu, memory, or storage resource requests and/or limits for pods and volumes.
See [default-resource-limits.yaml][default-resource-limits] for an example of setting resource request and limits for each component.

For more reading on tuning metering, use the [Resource Tuning documentation][tuning].

## Node Selectors

If you want to run the metering components on specific sets of nodes then you can set nodeSelectors on each component to control where each component of metering is scheduled to.
See [node-selectors.yaml][node-selectors-config] for an example of setting node selectors for each component.

## Image repositories and tags

You can override the image repositories and versions to test pre-releases or to deploy an image built by our CI for PRs or testing.
See [latest-versions.yaml][latest-versions] for an example of setting the repository and image tag for each component to use.

[latest-versions]: ../manifests/metering-config/latest-versions.yaml
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
[node-selectors-config]: ../manifests/metering-config/custom-node-selectors.yaml
[default-resource-limits]: ../manifests/metering-config/default-resource-limits.yaml
[tuning]: tuning.md

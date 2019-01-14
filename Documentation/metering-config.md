# Metering Configuration

Metering supports configuration options which may be set in the `spec` section of the `Metering` resource.

A minimal configuration example that doesn't override anything can be found in [default.yaml](../manifests/metering-config/default.yaml).

For details on different types of configuration read the relevant document:

- [common configuration options](common-configuration.md)
  - [pod resource requests and limits](common-configuration.md#resource-requests-and-limits)
  - [node selectors](common-configuration.md#node-selectors)
  - [image repositories and tags](common-configuration.md#image-repositories-and-tags)
- [configuring reporting-operator](configuring-reporting-operator.md)
  - [set the Prometheus URL](configuring-reporting-operator.md#prometheus-url)
  - [exposing the reporting API](configuring-reporting-operator.md#exposing-the-reporting-api)
  - [configuring Authentication on Openshift](configuring-reporting-operator.md#openshift-authentication)
- [configuring storage](configuring-storage.md)
  - [storing data in s3](configuring-storage.md#storing-data-in-s3)
- [configuring the Hive metastore](configuring-hive-metastore.md)
- [configuring aws billing correlation for cost correlation](configuring-aws-billing.md)

## Documentation conventions

This document and other documents in the operator-metering project follow the convention of describing nested fields in configuration settings using dots as separators.
For example:

```
spec.reporting-operator.spec.config.awsAccessKeyID
```

Refers to the following YAML structure and value:

```
spec:
  reporting-operator:
    spec:
      config:
        awsAccessKeyID: "REPLACEME"
```

## Using a custom configuration

To install the custom configuration file, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-config/default.yaml
```

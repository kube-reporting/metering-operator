# Metering Configuration

Metering supports configuration options which may be set in the `spec` section of the `MeteringConfig` resource.

**Note**: Metering does not support any default storage at this moment. If a storage configuration is not specified or improperly set, then it will fail the `MeteringConfig` schema validation.

For details on different types of configuration read the relevant document:

- [common configuration options](common-configuration.md)
  - [pod resource requests and limits](common-configuration.md#resource-requests-and-limits)
  - [node selectors](common-configuration.md#node-selectors)
  - [image repositories and tags](common-configuration.md#image-repositories-and-tags)
- [configuring reporting-operator](configuring-reporting-operator.md)
  - [set Prometheus connection configuration](configuring-reporting-operator.md#prometheus-connection)
  - [exposing the reporting API](configuring-reporting-operator.md#exposing-the-reporting-api)
  - [configuring Authentication on Openshift](configuring-reporting-operator.md#openshift-authentication)
- [configuring storage](configuring-storage.md)
  - [storing data in Amazon S3](configuring-storage.md#storing-data-in-amazon-s3)
- [configuring the Hive metastore](configuring-hive-metastore.md)
- [configuring aws billing correlation for cost correlation](configuring-aws-billing.md)

## Documentation conventions

This document and other documents in the metering-operator project follow the convention of describing nested fields in configuration settings using dots as separators.
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

**Note**: Ensure the environment variable `$METERING_NAMESPACE` is properly set to the correct namespace.

To install the custom configuration file, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-config/default.yaml
```

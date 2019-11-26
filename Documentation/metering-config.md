# Metering Con***REMOVED***guration

Metering supports con***REMOVED***guration options which may be set in the `spec` section of the `MeteringCon***REMOVED***g` resource.

**Note**: Metering does not support any default storage at this moment. If a storage con***REMOVED***guration is not speci***REMOVED***ed or improperly set, then it will fail the `MeteringCon***REMOVED***g` schema validation.

For details on different types of con***REMOVED***guration read the relevant document:

- [common con***REMOVED***guration options](common-con***REMOVED***guration.md)
  - [pod resource requests and limits](common-con***REMOVED***guration.md#resource-requests-and-limits)
  - [node selectors](common-con***REMOVED***guration.md#node-selectors)
  - [image repositories and tags](common-con***REMOVED***guration.md#image-repositories-and-tags)
- [con***REMOVED***guring reporting-operator](con***REMOVED***guring-reporting-operator.md)
  - [set Prometheus connection con***REMOVED***guration](con***REMOVED***guring-reporting-operator.md#prometheus-connection)
  - [exposing the reporting API](con***REMOVED***guring-reporting-operator.md#exposing-the-reporting-api)
  - [con***REMOVED***guring Authentication on Openshift](con***REMOVED***guring-reporting-operator.md#openshift-authentication)
- [con***REMOVED***guring storage](con***REMOVED***guring-storage.md)
  - [storing data in Amazon S3](con***REMOVED***guring-storage.md#storing-data-in-amazon-s3)
- [con***REMOVED***guring the Hive metastore](con***REMOVED***guring-hive-metastore.md)
- [con***REMOVED***guring aws billing correlation for cost correlation](con***REMOVED***guring-aws-billing.md)

## Documentation conventions

This document and other documents in the operator-metering project follow the convention of describing nested ***REMOVED***elds in con***REMOVED***guration settings using dots as separators.
For example:

```
spec.reporting-operator.spec.con***REMOVED***g.awsAccessKeyID
```

Refers to the following YAML structure and value:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        awsAccessKeyID: "REPLACEME"
```

## Using a custom con***REMOVED***guration

**Note**: Ensure the environment variable `$METERING_NAMESPACE` is properly set to the correct namespace.

To install the custom con***REMOVED***guration ***REMOVED***le, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-con***REMOVED***g/default.yaml
```

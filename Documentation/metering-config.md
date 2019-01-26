# Metering Con***REMOVED***guration

Metering supports con***REMOVED***guration options which may be set in the `spec` section of the `Metering` resource.

A minimal con***REMOVED***guration example that doesn't override anything can be found in [default.yaml](../manifests/metering-con***REMOVED***g/default.yaml).

For details on different types of con***REMOVED***guration read the relevant document:

- [common con***REMOVED***guration options](common-con***REMOVED***guration.md)
  - [pod resource requests and limits](common-con***REMOVED***guration.md#resource-requests-and-limits)
  - [node selectors](common-con***REMOVED***guration.md#node-selectors)
  - [image repositories and tags](common-con***REMOVED***guration.md#image-repositories-and-tags)
- [con***REMOVED***guring reporting-operator](con***REMOVED***guring-reporting-operator.md)
  - [set the Prometheus URL](con***REMOVED***guring-reporting-operator.md#prometheus-url)
  - [exposing the reporting API](con***REMOVED***guring-reporting-operator.md#exposing-the-reporting-api)
  - [con***REMOVED***guring Authentication on Openshift](con***REMOVED***guring-reporting-operator.md#openshift-authentication)
- [con***REMOVED***guring storage](con***REMOVED***guring-storage.md)
  - [storing data in s3](con***REMOVED***guring-storage.md#storing-data-in-s3)
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

To install the custom con***REMOVED***guration ***REMOVED***le, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-con***REMOVED***g/default.yaml
```

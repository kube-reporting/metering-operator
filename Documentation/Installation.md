# Installing Chargeback

Chargeback consists of a few components:

- A chargeback pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the chargeback pod to perform queries on the
  collected usage data.

## Prerequisites

In order to install and use chargeback the following components will be
necessary:

- A tectonic installed Kubernetes cluster, of version 1.8.0 or greater, or with
  a Tectonic Prometheus Operator to be of version 1.6.0 or greater (Prometheus
  operator v0.13).
- A properly con***REMOVED***gured kubectl to access the Kubernetes cluster.

To alter the version of the Tectonic Prometheus operator to be 1.6.0, run the
following command:

```
kubectl -n tectonic-system patch deploy tectonic-prometheus-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"tectonic-prometheus-operator","image":"quay.io/coreos/tectonic-prometheus-operator:v1.6.0"}]}}}}'
```

Once the operator changes the version of the `kube-state-metrics` pod to 1.0.1,
chargeback installation may proceed.

## Modifying default values

Chargeback will install into an existing namespace. Without con***REMOVED***guration, the
default is currently `team-chargeback`.

Chargeback also assumes it needs a docker pull secret to pull images, which
defaults to a secret named `coreos-pull-secret` in the `tectonic-system`
namespace.

To change either of these, override the following environment variables
(defaults are used in the example):

```
export CHARGEBACK_NAMESPACE=team-chargeback
export PULL_SECRET_NAMESPACE=tectonic-system
export PULL_SECRET=coreos-pull-secret
```

## Prometheus location

If Prometheus was setup by Tectonic and is running within the tectonic-system
namespace, then you can skip this section.

If you're running the Prometheus operator yourself (not using the Tectonic one),
then you need to con***REMOVED***gure the `prometheus-url` in
`manifests/chargeback/chargeback-con***REMOVED***g.yaml` to match the service created by
your Prometheus operator.

## Storing data in S3

By default the data that chargeback collects and generates is ephemeral, and
will not survive restarts of the hive pod it deploys. To make this data
persistent by storing it in S3, follow the instructions in the [storing data in
S3 document][Storing-Data-In-S3.md] before proceeding with these instructions.

## Run the install script

Run `./hack/install.sh` to install Chargeback on the cluster.

## Verifying operation

Check the logs of the "chargeback" deployment, there should be no errors:

```
kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Generating reports

With Chargeback now successfully installed, reports may be generated. Note that
with the default con***REMOVED***guration, chargeback will need to run for some time for
enough usage data to be built up to generate a report. Reports can be generated
by creating report objects in Kubernetes in the same namespace as Chargeback.
Some examples of report objects exist in the
`manifests/custom-resources/reports` directory.

To deploy an example pod usage by memory report, create the report in
Kubernetes:

```
kubectl -n $CHARGEBACK_NAMESPACE create -f manifests/custom-resources/reports/pod-memory-usage-by-node.yaml
```

Existing reports can be viewed in Kubernetes with the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE get reports
```

A report's status can be inspected by viewing the object with the `-o json`
flag:

```
kubectl -n $CHARGEBACK_NAMESPACE get report pod-memory-usage -o json
```

## Viewing reports

Once a report is ***REMOVED***nished the results can be fetched using an HTTP API available
via the chargeback pod.

First, set up a port forward via `kubectl` to the pod:

```
kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE port-forward {} 8080
```

And then `curl` can be used (or a web browser) to access ***REMOVED***nished reports by
name:

```
curl "localhost:8080/api/v1/reports/get?name=pod-memory-usage-by-node&format=csv"
```

The `name` parameter in the URL can be any report that is in the ***REMOVED***nished state,
and the `format` parameter can be either `csv` or `json`.

## Uninstall

To uninstall chargeback run:
```
./hack/uninstall.sh
```

## AWS Billing data setup

**AWS billing reports were temporarily removed from chargeback due to a
refactor, the following documentation is left in for when this functionality is
restored**

* Setup hourly billing reports in the AWS console by following [these](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html) instructions. Be sure to note the bucket, report pre***REMOVED***x, and report name speci***REMOVED***ed here.

* Create AWS access key with permissions for the bucket given above. The required permissions are:
```
s3:DeleteObject
s3:GetObject
s3:GetObjectAcl1
s3:PutObject
s3:PutObjectAcl
s3:GetBucketAcl
s3:ListBucket
s3:GetBucketLocation
```

Once you have an `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` refer to
[Set AWS Credentials](set-aws-credentials) and [Set AWS region](set-aws-region) for con***REMOVED***guring.

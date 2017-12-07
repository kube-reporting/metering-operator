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
$ kubectl -n tectonic-system patch deploy tectonic-prometheus-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"tectonic-prometheus-operator","image":"quay.io/coreos/tectonic-prometheus-operator:v1.6.0"}]}}}}'
```

Once the operator changes the version of the `kube-state-metrics` pod to 1.0.1,
chargeback installation may proceed.

## Installation

To install Chargeback you can run our installation script.
Before running the script, you can customize the installation if you want to
customize where Chargeback is installed, or if you want to change where it
stores data, etc.

### Modifying default values

Chargeback will install into an existing namespace. Without con***REMOVED***guration, the
default is currently `chargeback`.

Chargeback also assumes it needs a docker pull secret to pull images, which
defaults to a secret named `coreos-pull-secret` in the `tectonic-system`
namespace.

To change either of these, override the following environment variables
(defaults are used in the example):

```
$ export CHARGEBACK_NAMESPACE=chargeback
$ export PULL_SECRET_NAMESPACE=tectonic-system
$ export PULL_SECRET=coreos-pull-secret
```

### Prometheus location

If Prometheus was setup by Tectonic and is running within the tectonic-system
namespace, then you can skip this section.

If you're running the Prometheus operator yourself (not using the Tectonic one),
then you need to con***REMOVED***gure the `prometheus-url` in
`manifests/chargeback/chargeback-con***REMOVED***g.yaml` to match the service created by
your Prometheus operator.

### Storing data in S3

By default the data that chargeback collects and generates is ephemeral, and
will not survive restarts of the hive pod it deploys. To make this data
persistent by storing it in S3, follow the instructions in the [storing data in
S3 document](Storing-Data-In-S3.md) before proceeding with these instructions.

### AWS Billing Correlation

Chargeback is able to correlate cluster usage information with [AWS detailed
billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by setting the bucket and
pre***REMOVED***x in `manifests/custom-resources/datastores/aws-billing.yaml` to match the
location billing reports are con***REMOVED***gured to be stored, and having the
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables set with
credentials capable of accessing the bucket when the install command is run.

If the AWS Cost and Usage Reports are not enabled on your account, instructions
to enable them can be found [here][enable-aws-billing].

### Run the install script

Chargeback can now be installed with the following command:

```
$ ./hack/install.sh
```

### Uninstall

If chargeback has been installed manually, it can be uninstalled at any point by
running the following command:

```
$ ./hack/uninstall.sh
```

## Verifying operation

Check the logs of the "chargeback" deployment, there should be no errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using chargeback

For instructions on using chargeback, please read the documentation on [using
chargeback](Using-chargeback.md)

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html

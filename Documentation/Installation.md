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
- A properly configured kubectl to access the Kubernetes cluster.

To alter the version of the Tectonic Prometheus operator to be 1.6.0, run the
following command:

```
$ kubectl -n tectonic-system patch deploy tectonic-prometheus-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"tectonic-prometheus-operator","image":"quay.io/coreos/tectonic-prometheus-operator:v1.6.0"}]}}}}'
```

Once the operator changes the version of the `kube-state-metrics` pod to 1.0.1,
chargeback installation may proceed.

## Installing via ALM

If ALM is installed on the cluster it can be used to deploy chargeback. At this
point ALM will not handle upgrading or uninstalling chargeback, but these
features should become available in the future.

ALM will install chargeback into the `tectonic-system` namespace, and thus
access to this namespace is required to interact with chargeback.

### Creating CRDs

ALM expects the requisite CRDs for an application to be created before it will
install the app. Currently this should be done manually, with the following
command:

```
$ kubectl create -f manifests/custom-resource-definitions
```

### Installing chargeback

With the CRDs created, the cluster service version file can now be created to
instruct ALM to install chargeback:

```
$ kubectl -n tectonic-system create -f manifests/alm/chargeback.clusterserviceversion.yaml
```

To be able to follow along with the rest of the installation document, set
`CHARGEBACK_NAMESPACE` to `tectonic-system`:

```
$ export CHARGEBACK_NAMESPACE=tectonic-system
```

And proceed to [Verifying operation](#Verifying operation)

## Installing manually

If ALM is unavailable or chargeback is to be installed into a namespace other
than `tectonic-system` it can be installed by hand.

### Modifying default values

Chargeback will install into an existing namespace. Without configuration, the
default is currently `team-chargeback`.

Chargeback also assumes it needs a docker pull secret to pull images, which
defaults to a secret named `coreos-pull-secret` in the `tectonic-system`
namespace.

To change either of these, override the following environment variables
(defaults are used in the example):

```
$ export CHARGEBACK_NAMESPACE=team-chargeback
$ export PULL_SECRET_NAMESPACE=tectonic-system
$ export PULL_SECRET=coreos-pull-secret
```

### Prometheus location

If Prometheus was setup by Tectonic and is running within the tectonic-system
namespace, then you can skip this section.

If you're running the Prometheus operator yourself (not using the Tectonic one),
then you need to configure the `prometheus-url` in
`manifests/chargeback/chargeback-config.yaml` to match the service created by
your Prometheus operator.

### Storing data in S3

By default the data that chargeback collects and generates is ephemeral, and
will not survive restarts of the hive pod it deploys. To make this data
persistent by storing it in S3, follow the instructions in the [storing data in
S3 document][Storing-Data-In-S3.md] before proceeding with these instructions.

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

### AWS Billing data setup

**AWS billing reports were temporarily removed from chargeback due to a
refactor, the following documentation is left in for when this functionality is
restored**

* Setup hourly billing reports in the AWS console by following [these](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html) instructions. Be sure to note the bucket, report prefix, and report name specified here.

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
[Set AWS Credentials](set-aws-credentials) and [Set AWS region](set-aws-region) for configuring.

## Configuration

Chargeback supports a few configuration options which can currently be set by
creating a secret named `chargeback-settings` with a key of `values.yaml`.

An example configuration file can be found in
[manifests/chargeback-config/custom-values.yaml][example-config]
For details on customizing this file, read the
[common-configuration-options](common-configuration-options) section below.

### Using a custom configuration

To install the custom configuration file, you can run the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-file 'values.yaml=manifests/chargeback-config/custom-values.yaml'
```

However, this doesn't make it very easy to make changes without deleting and
re-creating the secret, so you can also use this one-liner which will do the
same thing, but will use `kubectl apply`

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-file 'values.yaml=manifests/chargeback-config/custom-values.yaml' -o yaml --dry-run > /tmp/chargeback-settings-secret.yaml && kubectl apply -n $CHARGEBACK_NAMESPACE -f /tmp/chargeback-settings-secret.yaml
```

## Common configuration options

The example manifest
[manifests/chargeback-config/custom-values.yaml][example-config]
contains has the most common configuration values a user might want to modify
included in it already, the options are commented out and can be uncommented
and adjusted as needed. The values are the defaults, and should work for most
users, but if you run into resource issues, you can increase the limits as
needed.

### Storing data in S3

By default the data that chargeback collects and generates is stored in a
single node HDFS cluster which is backed by a persistent volume. If you would
instead prefer to store the data in a location outside of the cluster, you can
configure Chargeback to store data in s3.

To use S3, for storage, uncomment the `defaultStorage:` section in the example
[manifests/chargeback-config/custom-values.yaml][example-config] configuration.
Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.config` and `presto.config`
sections.

Please note that this must be done before installation, and if it is changed
post-install you may see unexpected behavior.

### AWS Billing Correlation

Chargeback is able to correlate cluster usage information with [AWS detailed
billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by modifying the example
[manifests/chargeback-config/custom-values.yaml][example-config] configuration.

To enable AWS Billing correlation, please ensure the AWS Cost and Usage Reports
are enabled on your account, instructions to enable them can be found
[here][enable-aws-billing].

Next, in the example configuration manifest, you need to update the
`awsBillingDataSource:` section. Change the `enabled` value to `true`, and then
update the `bucket` and `prefix` to the location of your AWS Detailed billing
report.  Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.config` and `presto.config`
sections.

This can be done either pre-install or post-install. Note that disabling it
post-install can cause errors in the chargeback-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-config]: manifests/chargeback-config/custom-values.yaml

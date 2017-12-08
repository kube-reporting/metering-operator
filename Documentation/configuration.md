## Con***REMOVED***guration

Chargeback supports a few con***REMOVED***guration options which can currently be set by
creating a secret named `chargeback-settings` with a key of `values.yaml`.

An example con***REMOVED***guration ***REMOVED***le can be found in
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g]
For details on customizing this ***REMOVED***le, read the
[common-con***REMOVED***guration-options](common-con***REMOVED***guration-options) section below.

### Using a custom con***REMOVED***guration

To install the custom con***REMOVED***guration ***REMOVED***le, you can run the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml'
```

However, this doesn't make it very easy to make changes without deleting and
re-creating the secret, so you can also use this one-liner which will do the
same thing, but will use `kubectl apply`

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml' -o yaml --dry-run > /tmp/chargeback-settings-secret.yaml && kubectl apply -n $CHARGEBACK_NAMESPACE -f /tmp/chargeback-settings-secret.yaml
```

## Common con***REMOVED***guration options

The example manifest
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g]
contains has the most common con***REMOVED***guration values a user might want to modify
included in it already, the options are commented out and can be uncommented
and adjusted as needed. The values are the defaults, and should work for most
users, but if you run into resource issues, you can increase the limits as
needed.

### Storing data in S3

By default the data that chargeback collects and generates is stored in a
single node HDFS cluster which is backed by a persistent volume. If you would
instead prefer to store the data in a location outside of the cluster, you can
con***REMOVED***gure Chargeback to store data in s3.

To use S3, for storage, uncomment the `defaultStorage:` section in the example
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.
Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g`
sections.

Please note that this must be done before installation, and if it is changed
post-install you may see unexpected behavior.

### AWS Billing Correlation

Chargeback is able to correlate cluster usage information with [AWS detailed
billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by modifying the example
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.

To enable AWS Billing correlation, please ensure the AWS Cost and Usage Reports
are enabled on your account, instructions to enable them can be found
[here][enable-aws-billing].

Next, in the example con***REMOVED***guration manifest, you need to update the
`awsBillingDataSource:` section. Change the `enabled` value to `true`, and then
update the `bucket` and `pre***REMOVED***x` to the location of your AWS Detailed billing
report.  Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g`
sections.

This can be done either pre-install or post-install. Note that disabling it
post-install can cause errors in the chargeback-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-con***REMOVED***g]: manifests/chargeback-con***REMOVED***g/custom-values.yaml

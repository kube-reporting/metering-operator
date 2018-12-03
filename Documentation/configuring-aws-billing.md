# AWS billing correlation

Metering is able to correlate cluster usage information with [AWS detailed billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by modifying the example [aws-billing.yaml][example-con***REMOVED***g] con***REMOVED***guration.

To enable AWS billing correlation, ***REMOVED***rst ensure the AWS Cost and Usage Reports are enabled.
For more information, see [Turning on the AWS Cost and Usage report][enable-aws-billing] in the AWS documentation.

Next, update update the `bucket`, `pre***REMOVED***x` and `region` to the location of your AWS Detailed billing report in the `openshift-reporting.spec.awsBillingReportDataSource` in the [aws-billing.yaml][example-con***REMOVED***g] example con***REMOVED***guration manifest.

Then, set the `awsAccessKeyID` and `awsSecretAccessKey` in the `spec.reporting-operator.spec.con***REMOVED***g` and `spec.presto.spec.con***REMOVED***g` sections.

To retrieve data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-only.json](aws/read-only.json) ***REMOVED***le.
Replace `operator-metering-data` with the name of your bucket.

This can be done either pre-install or post-install. Note that disabling it post-install can cause errors in the reporting-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/aws-billing.yaml

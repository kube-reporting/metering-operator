# AWS billing correlation

Metering is able to correlate cluster usage information with [AWS detailed billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by modifying the example [aws-billing.yaml][example-config] configuration.

To enable AWS billing correlation, first ensure the AWS Cost and Usage Reports are enabled.
For more information, see [Turning on the AWS Cost and Usage report][enable-aws-billing] in the AWS documentation.

Next, update update the `bucket`, `prefix` and `region` to the location of your AWS Detailed billing report in the `openshift-reporting.spec.awsBillingReportDataSource` in the [aws-billing.yaml][example-config] example configuration manifest.

The `spec.reporting-operator.spec.config.aws.secretName` and `spec.presto.spec.config.aws.secretName` fields should be set to the name of a secret in the metering namespace containing AWS credentials in the `data.aws-access-key-id` and `data.aws-secret-access-key` fields.

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: your-aws-secret
data:
  aws-access-key-id: "dGVzdAo="
  aws-secret-access-key: "c2VjcmV0Cg=="
```

To store data in S3, the `aws-access-key-id` and `aws-secret-access-key` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-write.json](aws/read-write.json) file.
Replace `operator-metering-data` with the name of your bucket.

This can be done either pre-install or post-install. Note that disabling it post-install can cause errors in the reporting-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-config]: ../manifests/metering-config/aws-billing.yaml

# Installing Chargeback

Chargeback consists of two components: a daemon which retrieves usage information from Prometheus and a operator which manages Hive and Presto clusters to perform queries on the collected usage data.

## AWS Billing data setup
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

* Create secret with correct credentials for the bucket above and deploy to cluster:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws
  namespace: tectonic-chargeback
type: Opaque
data:
  AWS_ACCESS_KEY_ID: <base64 encoded ID>
  AWS_SECRET_ACCESS_KEY: <base64 encoded Secret>
```

* Modify **manifests/chargeback/chargeback.yaml** with the correct AWS region for the bucket.

## Pod usage data setup

* Create a S3 bucket to hold Pod usage data collected from the cluster. Currently, this bucket must be accessible with the same AWS credentials as the AWS billing report.

* Modify `S3_BUCKET` and `S3_PATH` in **manifests/promsum/promsum.yaml** to a valid bucket and pre***REMOVED***x for usage data to be stored.

## Install components
To install chargeback to it's own namespace, run:
```
./install.sh
```

### Verifying operation
Check the logs of the "promsum" deployment. There should be no errors and you should see billing records printed to stdout.

# Uninstall
To uninstall chargeback run:
```
./uninstall.sh
```

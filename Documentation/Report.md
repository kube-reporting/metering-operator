# Reports
Chargeback produces reports derived from usage data sources which can be used in further analysis and filtering. The `Report` custom Kubernetes resource is used to manage the execution and status of reports.

## Execution time
Reports take a variable amount of time to complete and can potentially need to run for very long periods.

The duration a report takes to runn is determined by:
* report type
* amount of data being analyzed
* system performance (memory, CPU)
* network performance

## Report Object
A single `Report` resource corresponds to a specific run of a report. Once the object is created Chargeback starts collecting and analyzing the data required to perform the report. A report cannot be updated after it's created and currently must run to completion.

### Time format
Instances of a timestamps should be [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) encoded. Times with local offsets will be converted to UTC.

### S3 bucket
Chargeback uses S3 buckets to collect data and write reports after it's been analyzed. The location is given as the `bucket` and the `prefix` of keys where data is stored.

*Example*
```yaml
bucket: east-region-clusters
prefix: july-data
```

### Spec
The type of report that is performed and what data is used is determined by the `spec` section of the report object. There is configuration properties shared by all reports and others that are specific to a report type.

#### Period
The period that a report is generated for is specified by the interval between `reportingStart` and `reportingEnd`.

*Example*
```yaml
reportingStart: '2017-07-02T00:00:00Z'
reportingEnd: '2017-07-29T00:00:00Z'
```

#### Output
The result of a report is stored in the S3 bucket given by the `output` field.

*Example*
```yaml
output:
  bucket: usage-reports
  prefix: east
```

### Report types
Chargeback offers multiple different views of data that can be compiled in a report. Only a single type can be run per `Report` object, however multiple may be supported in the future.

#### Pod usage
This report returns total memory [requested](https://kubernetes.io/docs/api-reference/v1.7/#resourcerequirements-v1-core) by Pods during the given period. The unit of usage is Byte Second, which represents the request of 1 byte over a second.

This report is run by specifying an S3 bucket in the `chargeback` field. The values here should match the `S3_BUCKET` and `S3_PATH` in the Deployment of **manifests/promsum/promsum.yaml** created in the cluster being reported on.

*Example*
```yaml
apiVersion: chargeback.coreos.com/prealpha
kind: Report
metadata:
  name: pods
  spec:
    reportingStart: '2017-07-02T00:00:00Z'
    reportingEnd: '2017-07-29T00:00:00Z'
  chargeback:
    bucket: <INSERT BUCKET FROM PROMSUM>
    prefix: <INSERT S3 PREFIX FROM PROMSUM>
  output:
    bucket: <OUTPUT BUCKET>
    prefix: <OUTPUT PREFIX>
```

*Report columns*

The report contains the following columns sequentially:
1. Pod name
1. Namespace
1. Node name
1. Usage (Byte * Second)
1. Time first seen Pod
1. Time last seen Pod
1. Pod labels (stored in JSON map)

#### AWS Pod Cost
This report determines the cost of running a Pod over a given period by rating the amount of memory requested by a Pod (same as Pod usage) against AWS billing data. This gives a measure of the cost of operating specific software on Kubernetes.

This report is run by specifying the `chargeback` field (using instructions above) and a bucket for the AWS usage data in `aws`. The bucket should be the one specified when creating the usage report in the AWS console. The bucket `prefix` should take the form "<AWS Report Prefix>/<AWS Report Name>" based on the values entered when creating the AWS report.

*Example*
```yaml
apiVersion: chargeback.coreos.com/prealpha
kind: Report
metadata:
  name: pods
  spec:
    reportingStart: '2017-07-02T00:00:00Z'
    reportingEnd: '2017-07-29T00:00:00Z'
  aws:
    bucket: <AWS Report bucket>
    prefix: <AWS Report Prefix>/<AWS Report Name>
  output:
    bucket: <OUTPUT BUCKET>
    prefix: <OUTPUT PREFIX>
```

*Report columns*

The report contains the following columns sequentially:
1. Pod name
1. Namespace
1. Node name
1. Cost (US dollars)
1. Time first seen Pod
1. Time last seen Pod
1. Pod labels (stored in JSON map)

### Status
The execution of a `Report` can be tracked using it's status field. Any errors occurring during the preparation of a report will be recorded here.

#### Phase
A report can have the following states:
* `Waiting` - Acknowledgment of the request for a report. There are no problems reported yet but the `output` field may provide the reason for the delay. The `spec` can still be modified.
* `Started` - Chargeback has started executing the report. No modifications can be made at this point.
* `Finished` - The report successfully completed execution.
* `Error` - A failure occurred running the report. Details are provided in the `output` field.

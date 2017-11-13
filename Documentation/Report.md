# Reports
Chargeback produces reports derived from usage data sources which can be used in further analysis and filtering. The `Report` custom Kubernetes resource is used to manage the execution and status of reports.

## Execution time
Reports take a variable amount of time to complete and can potentially need to run for very long periods.

The duration a report takes to run is determined by:
* report type
* amount of data being analyzed
* system performance (memory, CPU)
* network performance

## Report Object
A single `Report` resource corresponds to a specific run of a report. Once the object is created Chargeback starts analyzing the data required to perform the report. A report cannot be updated after it's created and currently must run to completion.

### Time format
Instances of a timestamps should be [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) encoded. Times with local offsets will be converted to UTC.

### S3 bucket
Chargeback uses S3 buckets to collect data and write reports after it's been analyzed. The location is given as the `bucket` and the `prefix` of keys where data is stored.

*Example*
```yaml
bucket: east-region-clusters
prefix: july-data
```

#### Period
The period that a report is generated for is specified by the interval between `reportingStart` and `reportingEnd`.

*Example*
```yaml
reportingStart: '2017-07-02T00:00:00Z'
reportingEnd: '2017-07-29T00:00:00Z'
```

#### Running reports immediately
Reports will by default wait until 5 minutes after `reportingEnd`. This 5 minute grace period can be configured with the `gracePeriod` field, and reports can be set to run immediately regardless of the end time with the `runImmediately` flag.

*Example*
```
runImmediately: true
```

*Example*
```
gracePeriod: 30s
```

#### Output
The result of a report is stored in the S3 bucket given by the `output` field.

*Example*
```yaml
output:
  bucket: usage-reports
  prefix: east
```

### Status
The execution of a `Report` can be tracked using it's status field. Any errors occurring during the preparation of a report will be recorded here.

#### Phase
A report can have the following states:
* `Started` - Chargeback has started executing the report. No modifications can be made at this point.
* `Finished` - The report successfully completed execution.
* `Error` - A failure occurred running the report. Details are provided in the `output` field.

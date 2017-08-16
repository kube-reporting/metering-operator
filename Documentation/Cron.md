# Cron
Chargeback allows reports to be scheduled to run at regular intervals. This can be useful for usage reporting automation. The schedules used to run reports is defined using the `Cron` custom Kubernetes resource.

## Cron Object
Scheduled running of reports is initialized by the creation of a `Cron` object. The schedule can be cancelled by deleting the object.

**Example**
```yaml
apiVersion: cron.coreos.com/v1
kind: Cron
metadata:
  name: hourly-aws
spec:
  suspend: false
  frequency: Hourly
  reportTemplate:
    metadata:
      generateName: hourly
    spec:
      chargeback:
        bucket: <Promsum bucket>
        prefix: <Promsum bucket prefix>
      aws:
        bucket: <AWS report data bucket>
        prefix: <AWS report prefix>
      output:
        bucket: <Output bucket>
        prefix: <Output prefix>
```

## CronSpec
The `spec` field defines the configuration of a report and the frequency that it's run.

### Frequency
The schedule that is used to run reports is determined by the predefined frequency chosen by the user.

The following options are currently available:
* `Hourly` - Report every hour at :00
* `Daily` - Report for every UTC day
* `Weekly`

**Example**
```yaml
frequency: Weekly
```

#### Offset
All scheduled reports introduce an offset of **16 hours** to allow cloud provider data to be created. This may be more customizable in the future.

## Suspend
Report generation can be temporarily paused by marking the `suspend` field.

**Example**
```yaml
suspend: true
```

### ReportTemplate
The configuration of the report gets that gets run is configured by `reportTemplate`.

**Example**
```yaml
reportTemplate:
  metadata:
    generateName: hourly
  spec:
    chargeback:
      bucket: <Promsum bucket>
      prefix: <Promsum bucket prefix>
    aws:
      bucket: <AWS report data bucket>
      prefix: <AWS report prefix>
    output:
      bucket: <Output bucket>
      prefix: <Output prefix>
```

#### ObjectMeta
The `meta` field defines the [ObjectMeta](https://kubernetes.io/docs/api-reference/v1.7/#objectmeta-v1-meta) that `Report` objects are created with.

It's recommended that the `generateName` field be used to avoid naming conflicts.

#### ReportSpec
The type of report that's run and it's options are defined by `spec`. The `Report` documentation provides the details of how this can be configured.

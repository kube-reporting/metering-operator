# Storing Data in S3

Chargeback supports storing collected usage information and generated reports in
S3. Data stored locally in the cluster (which is the default) will not survive
restarts of the hive pod. By configuring chargeback to store data in S3, this
data will become persistent.

Some resources in Kubernetes must be edited to use S3. These edits can be done
with the Tectonic console, but this document will assume all edits are being
done via `kubectl`.

## Set AWS Credentials

There is a secret named `chargeback-secrets` in chargeback's namespace that must
be populated with AWS credentials. Chargeback will use these credentials to
store data in S3.

There is a script available to help with setting these credentials. Make sure
`CHARGEBACK_NAMESPACE` is set to the right namespace, and then set the
environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` to the
desired credentials and run:

```
./hack/update-aws-credentials.sh
```

After the secret has been updated, delete the relevant chargeback pods so they
can be recreated with the new secret:

```
kubectl -n CHARGEBACK_NAMESPACE delete pod -l app=presto
kubectl -n CHARGEBACK_NAMESPACE delete pod -l app=hive
kubectl -n $CHARGEBACK_NAMESPACE delete pod -l app=chargeback
```

## Set AWS region

Chargeback currently needs to be configured to use a specific AWS region. Edit
chargeback's config with the following command, and change the `aws-region`
value to match the region of the S3 bucket that is to be used. Then, delete the
chargeback pod so it can be recreated with the new config.

```
kubectl -n $CHARGEBACK_NAMESPACE edit configmap chargeback-config
kubectl -n $CHARGEBACK_NAMESPACE delete pod -l app=chargeback
```

## Modify Chargeback data stores

Chargeback has a few CRDs that are installed to control what types of reports
can be generated. More information on this is available in the [documentation on
Chargeback's CRD model][crd-model], but in summary the data store CRDs describe
where data for a given Prometheus query should be stored. These data stores can
be modified to point to a S3 bucket instead of local storage.

For each data store with a `promsum` section, replace:

```
storage:
  local: {}
  s3: null
```

with:

```
storage:
  local: null
  s3:
    bucket: MY-BUCKET-NAME
    prefix: MY-PREFIX
```

Existing data stores can be viewed with the command:

```
kubectl -n $CHARGEBACK_NAMESPACE get reportdatastores
```

And a given data store can be modified with the command:

```
kubectl -n $CHARGEBACK_NAMESPACE edit reportdatastore [NAME]
```

As an example, here's the `pod-request-cpu-cores` data store after being
modified to store data in the `chargeback` bucket under the `promsum/cpu_by_pod`
prefix:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: ReportDataStore
metadata:
  name: "pod-request-cpu-cores"
  labels:
    tectonic-chargeback: "true"
spec:
  promsum:
    query: "pod-request-cpu-cores"
    storage:
      local: null
      s3:
        bucket: chargeback
        prefix: promsum/pod_request_cpu_cores
```

## Set an output location on reports

Reports also must specify to put report results in S3 if they shouldn't be
stored locally in the cluster. To alter the example reports in
`manifests/custom-resources/reports` to store data in S3, replace:

```
output:
  local: {}
```

with:

```
output:
  s3:
    bucket: MY-BUCKET-NAME
    prefix: MY-PREFIX
```

[crd-model]: CRD-Model.md

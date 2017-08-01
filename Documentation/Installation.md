# Installing Chargeback

Chargeback consists of two components: a daemon which retrieves usage information from Prometheus and a operator which manages Hive and Presto clusters to perform queries on the collected usage data.

## Scraper installation

* Modify `S3_BUCKET` and `S3_PATH` in **manifests/promsum/promsum.yaml** to a valid bucket and prefix for usage data to be stored.

* Install Chargeback manifests:
```
kubectl apply -f install/scraper
```
* Modify secret with correct credentials for the bucket above and deploy to cluster:
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

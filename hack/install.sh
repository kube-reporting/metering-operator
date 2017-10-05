#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

accessKey=${AWS_ACCESS_KEY_ID-"<base64 encoded AWS_ACCESS_KEY_ID>"}
accessSecret=${AWS_SECRET_ACCESS_KEY-"<base64 encoded AWS_SECRET_ACCESS_KEY>"}
setupAWS=n
if [[ $accessKey != \<b* ]] && [[ $accessKey != \<b* ]]; then
  if [ -t 0 ] && [ -t 1 ]; then
    read -p "AWS credentials (${accessKey}) detected. Would you like to create a secret for Chargeback using them? [y/N]: " setupAWS
  ***REMOVED***
***REMOVED***

if [[ "${setupAWS}" == "y" ]]; then
  sed \
      -e 's/aws-access-key-id: "REPLACEME"/aws-access-key-id: "'$(echo -n ${accessKey} | base64)'"/g' \
      -e 's/aws-secret-access-key: "REPLACEME"/aws-secret-access-key: "'$(echo -n ${accessSecret} | base64)'"/g' \
      manifests/chargeback/chargeback-secrets.yaml.dist \
      > manifests/chargeback/chargeback-secrets.yaml
***REMOVED***
  echo "To have chargeback setup AWS credentials for you: set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY, and re-run this script."
  echo "Alternatively, you can manually create the secret by copying the manifest manifests/chargeback/chargeback-secrets.yaml.dist to: man***REMOVED***ests/chargeback/chargeback-secrets.yaml and updating it with your credentials."
***REMOVED***

msg "Con***REMOVED***guring pull secrets"
copy-tectonic-pull

msg "Installing query and collection layer"
kube-install manifests/hive manifests/presto manifests/chargeback

msg "Populating chargeback CRDs"
kube-install manifests/chargeback-resources


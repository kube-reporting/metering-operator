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
  sed \
      -e 's/aws-access-key-id: "REPLACEME"/aws-access-key-id: ""/g' \
      -e 's/aws-secret-access-key: "REPLACEME"/aws-secret-access-key: ""/g' \
      manifests/chargeback/chargeback-secrets.yaml.dist \
      > manifests/chargeback/chargeback-secrets.yaml
***REMOVED***

msg "Con***REMOVED***guring pull secrets"
copy-tectonic-pull

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    manifests/custom-resource-de***REMOVED***nitons

msg "Installing query and collection layer"
kube-install \
    manifests/hive \
    manifests/presto \
    manifests/chargeback


msg "Installing Custom Resources"
kube-install \
    manifests/custom-resources/prom-queries \
    manifests/custom-resources/datastores \
    manifests/custom-resources/report-queries


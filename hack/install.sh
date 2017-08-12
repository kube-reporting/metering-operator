#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Creating chargeback namespace"
kube-install manifests/chargeback/namespace.yaml

accessKey=${AWS_ACCESS_KEY_ID-"<base64 encoded AWS_ACCESS_KEY_ID>"}
accessSecret=${AWS_SECRET_ACCESS_KEY-"<base64 encoded AWS_SECRET_ACCESS_KEY>"}
setupAWS=n
if [[ $accessKey != \<b* ]] && [[ $accessKey != \<b* ]]; then
  if [ -t 0 ] && [ -t 1 ]; then
    read -p "AWS credentials (${accessKey}) detected. Would you like to create a secret for Chargeback using them? [y/N]" setupAWS
  ***REMOVED***
  accessKey=$(printf "${accessKey}" | base64)
  accessSecret=$(printf "${accessSecret}" | base64)
***REMOVED***

if [[ "${setupAWS}" == "y" ]]; then
  aws_secret ${accessKey} ${accessSecret} | kube-install -
***REMOVED***
  echo "To have chargeback setup AWS credentials for you: set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY, and re-run this script."
  echo "Alternatively, you can manually create the secret using the manifest below:"
  echo "-------"
  aws_secret "${accessKey}" "${accessSecret}"
  echo "-------"
***REMOVED***

msg "Con***REMOVED***guring pull secrets"
copy-tectonic-pull

msg "Installing collection layer (with build of kube-state-metrics with Node info)"
kube-install manifests/kube-state-metrics manifests/promsum manifests/prom-operator

msg "Installing query layer"
kube-install manifests/hive manifests/presto manifests/chargeback

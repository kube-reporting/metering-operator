#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

accessKey=${AWS_ACCESS_KEY_ID-"<base64 encoded AWS_ACCESS_KEY_ID>"}
accessSecret=${AWS_SECRET_ACCESS_KEY-"<base64 encoded AWS_SECRET_ACCESS_KEY>"}
setupAWS=n
if [[ $accessKey != \<b* ]] && [[ $accessKey != \<b* ]]; then
  if [ -t 0 ] && [ -t 1 ]; then
    read -p "AWS credentials (${accessKey}) detected. Would you like to create a secret for Chargeback using them? [y/N]" setupAWS
  fi
  accessKey=$(printf "${accessKey}" | base64)
  accessSecret=$(printf "${accessSecret}" | base64)
fi

if [[ "${setupAWS}" == "y" ]]; then
  aws_secret ${accessKey} ${accessSecret} | kube-install -
else
  echo "To have chargeback setup AWS credentials for you: set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY, and re-run this script."
  echo "Alternatively, you can manually create the secret using the manifest below:"
  echo "-------"
  aws_secret "${accessKey}" "${accessSecret}"
  echo "-------"
fi

msg "Configuring pull secrets"
copy-tectonic-pull

msg "Installing collection layer"
kube-install manifests/promsum

msg "Installing query layer"
kube-install manifests/hive manifests/presto manifests/chargeback

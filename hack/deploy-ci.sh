#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
: ${DEPLOY_TAG:?}

export CHARGEBACK_CR_FILE="/tmp/custom-chargeback-cr-${DEPLOY_TAG}.yaml"
export INSTALLER_MANIFEST_DIR="/tmp/installer_manifests-${DEPLOY_TAG}"
export DELETE_PVCS=true

: "${ENABLE_AWS_BILLING:=false}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"

cat <<EOF > "$CHARGEBACK_CR_FILE"
apiVersion: chargeback.coreos.com/v1alpha1
kind: Chargeback
metadata:
  name: "tectonic-chargeback"
spec:
  chargeback-operator:
    image:
      tag: ${DEPLOY_TAG}

    con***REMOVED***g:
      disablePromsum: true
      awsBillingDataSource:
        enabled: ${ENABLE_AWS_BILLING}
        bucket: "${AWS_BILLING_BUCKET}"
        pre***REMOVED***x: "${AWS_BILLING_BUCKET_PREFIX}"
      awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
      awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"

  presto:
    con***REMOVED***g:
      awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
      awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"
    presto:
      image:
        tag: ${DEPLOY_TAG}
    hive:
      image:
        tag: ${DEPLOY_TAG}

  hdfs:
    image:
      tag: ${DEPLOY_TAG}
EOF

CUSTOM_VALUES_FILE="/tmp/helm-operator-values-${DEPLOY_TAG}.yaml"

cat <<EOF > "$CUSTOM_VALUES_FILE"
name: chargeback-helm-operator
operator:
  image:
    tag: ${DEPLOY_TAG}
EOF

echo "Creating installer manifests"

./hack/create-installer-manifests.sh "$CUSTOM_VALUES_FILE"

echo "Deploying"

./hack/deploy.sh

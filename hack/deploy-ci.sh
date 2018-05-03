#!/bin/bash
set -e

: "${DEPLOY_TAG:?}"
: "${DEPLOY_PLATFORM:?must be set to either tectonic, openshift, or generic}"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

TMP_DIR="$(mktemp -d)"
DEPLOY_DIR="$DIR/../manifests/deploy"

export METERING_CR_FILE=${METERING_CR_FILE:-"$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"}
export CUSTOM_DEPLOY_MANIFEST_DIR=${CUSTOM_DEPLOY_MANIFEST_DIR:-"$TMP_DIR/custom-deploy-manifests-${DEPLOY_TAG}"}
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}
export DELETE_PVCS=${DELETE_PVCS:-true}

: "${ENABLE_AWS_BILLING:=false}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"

cat <<EOF > "$METERING_CR_FILE"
apiVersion: chargeback.coreos.com/v1alpha1
kind: Metering
metadata:
  name: "operator-metering"
spec:
  metering-operator:
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
      terminationGracePeriodSeconds: 0
      image:
        tag: ${DEPLOY_TAG}
    hive:
      terminationGracePeriodSeconds: 0
      image:
        tag: ${DEPLOY_TAG}

  hdfs:
    image:
      tag: ${DEPLOY_TAG}
    datanode:
      terminationGracePeriodSeconds: 0
    namenode:
      terminationGracePeriodSeconds: 0
EOF

cat <<EOF > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
image:
  tag: ${DEPLOY_TAG}
reconcileIntervalSeconds: 5
EOF

cat <<EOF > "$CUSTOM_ALM_OVERRIDE_VALUES"
name: metering-helm-operator.v${DEPLOY_TAG}
spec:
  version: ${DEPLOY_TAG}
  labels:
    alm-status-descriptors: metering-helm-operator.v${DEPLOY_TAG}
    alm-owner-metering: metering-helm-operator
  matchLabels:
    alm-owner-metering: metering-helm-operator
EOF

echo "Creating metering manifests"
"$DIR/create-metering-manifests.sh" "$CUSTOM_DEPLOY_MANIFEST_DIR"

echo "Deploying"

export DEPLOY_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFEST_DIR"
./hack/deploy.sh

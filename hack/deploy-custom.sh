#!/bin/bash
set -e

: "${DEPLOY_TAG:?}"
: "${DEPLOY_PLATFORM:?must be set to either openshift, or generic}"

TMP_DIR="$(mktemp -d)"

export INSTALL_METHOD="${DEPLOY_PLATFORM}-direct"
export METERING_CR_FILE=${METERING_CR_FILE:-"$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"}
export CUSTOM_DEPLOY_MANIFESTS_DIR=${CUSTOM_DEPLOY_MANIFESTS_DIR:-"$TMP_DIR/custom-deploy-manifests-${DEPLOY_TAG}"}
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}
export DELETE_PVCS=${DELETE_PVCS:-true}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${ENABLE_AWS_BILLING:=false}"
: "${DISABLE_PROMSUM:=false}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"
: "${AWS_BILLING_BUCKET_REGION:=}"
: "${METERING_CREATE_PULL_SECRET:=false}"
: "${METERING_PULL_SECRET_NAME:=metering-pull-secret}"
: "${TERMINATION_GRACE_PERIOD_SECONDS:=0}"

: "${HDFS_NAMENODE_STORAGE_SIZE:=5Gi}"
: "${HDFS_DATANODE_STORAGE_SIZE:=5Gi}"

IMAGE_PULL_SECRET_TEXT=""
if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    IMAGE_PULL_SECRET_TEXT="imagePullSecrets: [ { name: \"$METERING_PULL_SECRET_NAME\" } ]"
fi

CUR_DATE="$(date +%s)"
DATE_ANNOTATION="\"metering.deploy-custom/deploy-time\": \"$CUR_DATE\""

cat <<EOF > "$METERING_CR_FILE"
apiVersion: metering.openshift.io/v1alpha1
kind: Metering
metadata:
  name: "${DEPLOY_PLATFORM}-metering"
spec:
  reporting-operator:
    spec:
      image:
        tag: ${DEPLOY_TAG}

      ${IMAGE_PULL_SECRET_TEXT:-}
      annotations: { $DATE_ANNOTATION }
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}

      config:
        disablePromsum: ${DISABLE_PROMSUM}
        awsBillingDataSource:
          enabled: ${ENABLE_AWS_BILLING}
          bucket: "${AWS_BILLING_BUCKET}"
          prefix: "${AWS_BILLING_BUCKET_PREFIX}"
          region: "${AWS_BILLING_BUCKET_REGION}"
        awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
        awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"


  presto:
    spec:
      ${IMAGE_PULL_SECRET_TEXT:-}
      config:
        awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
        awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"
      presto:
        annotations: { $DATE_ANNOTATION }
        terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
      hive:
        annotations: { $DATE_ANNOTATION }
        terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}

  hdfs:
    spec:
      ${IMAGE_PULL_SECRET_TEXT:-}
      datanode:
        annotations: { $DATE_ANNOTATION }
        terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
        storage:
          size: ${HDFS_DATANODE_STORAGE_SIZE}
      namenode:
        annotations: { $DATE_ANNOTATION }
        terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
        storage:
          size: ${HDFS_NAMENODE_STORAGE_SIZE}
EOF

cat <<EOF > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
image:
  tag: ${DEPLOY_TAG}
annotations: { $DATE_ANNOTATION }
reconcileIntervalSeconds: 5
${IMAGE_PULL_SECRET_TEXT:-}
EOF

cat <<EOF > "$CUSTOM_ALM_OVERRIDE_VALUES"
name: metering-operator.v${DEPLOY_TAG}
spec:
  version: ${DEPLOY_TAG}
  labels:
    alm-status-descriptors: metering-operator.v${DEPLOY_TAG}
    alm-owner-metering: metering-operator
  matchLabels:
    alm-owner-metering: metering-operator
EOF

echo "Creating metering manifests"
export MANIFEST_OUTPUT_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"$ROOT_DIR/hack/create-metering-manifests.sh"

echo "Deploying"
export DEPLOY_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"${ROOT_DIR}/hack/deploy.sh"

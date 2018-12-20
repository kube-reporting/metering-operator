#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")

: "${DEPLOY_TAG:?}"

# Used in deploy.sh
export DOCKER_USERNAME="$DOCKER_CREDS_USR"
export DOCKER_PASSWORD="$DOCKER_CREDS_PSW"

export DISABLE_PROMSUM=true
export METERING_CREATE_PULL_SECRET="${METERING_CREATE_PULL_SECRET:=true}"

export CUSTOM_METERING_CR_FILE="$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}

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
***REMOVED***

CUR_DATE="$(date +%s)"
DATE_ANNOTATION="\"metering.deploy-custom/deploy-time\": \"$CUR_DATE\""

cat <<EOF > "$CUSTOM_METERING_CR_FILE"
apiVersion: metering.openshift.io/v1alpha1
kind: Metering
metadata:
  name: "${DEPLOY_PLATFORM}-metering"
spec:
  openshift-reporting:
    spec:
      awsBillingReportDataSource:
        enabled: ${ENABLE_AWS_BILLING}
        bucket: "${AWS_BILLING_BUCKET}"
        pre***REMOVED***x: "${AWS_BILLING_BUCKET_PREFIX}"
        region: "${AWS_BILLING_BUCKET_REGION}"

  reporting-operator:
    spec:
      image:
        tag: ${DEPLOY_TAG}

      ${IMAGE_PULL_SECRET_TEXT:-}
      annotations: { $DATE_ANNOTATION }
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}

      con***REMOVED***g:
        disablePromsum: ${DISABLE_PROMSUM}
        awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
        awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"


  presto:
    spec:
      ${IMAGE_PULL_SECRET_TEXT:-}
      con***REMOVED***g:
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

export METERING_CR_FILE="$CUSTOM_METERING_CR_FILE"

"$DIR/deploy-custom.sh"

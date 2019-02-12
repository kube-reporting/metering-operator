#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")
ROOT_DIR="$DIR/.."
source "${ROOT_DIR}/hack/common.sh"

: "${DEPLOY_TAG:?}"

TMP_DIR="$(mktemp -d)"

# Used in deploy.sh
export DOCKER_USERNAME="$DOCKER_CREDS_USR"
export DOCKER_PASSWORD="$DOCKER_CREDS_PSW"

unset METERING_CR_FILE
export CUSTOM_METERING_CR_FILE="$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}

export METERING_PULL_SECRET_NAME
export METERING_CREATE_PULL_SECRET

: "${ENABLE_AWS_BILLING:=false}"
: "${DISABLE_PROMSUM:=true}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"
: "${AWS_BILLING_BUCKET_REGION:=}"
: "${METERING_CREATE_PULL_SECRET:=true}"
: "${METERING_PULL_SECRET_NAME:=metering-pull-secret}"
: "${TERMINATION_GRACE_PERIOD_SECONDS:=0}"
: "${HDFS_NAMENODE_STORAGE_SIZE:=5Gi}"
: "${HDFS_NAMENODE_MEMORY:=}"
: "${HDFS_DATANODE_STORAGE_SIZE:=5Gi}"
: "${HDFS_DATANODE_MEMORY:=}"
: "${HIVE_METASTORE_STORAGE_SIZE:=}"
: "${HIVE_METASTORE_MEMORY:=}"
: "${HIVE_METASTORE_CPU:=}"
: "${CUR_DATE:=$(date +%s)}"

HELM_ARGS=(\
    --set "deployTag=${DEPLOY_TAG}" \
    --set "enableAwsBilling=${ENABLE_AWS_BILLING}" \
    --set "disablePromsum=${DISABLE_PROMSUM}" \
    --set "awsAccessKeyId=${AWS_ACCESS_KEY_ID}" \
    --set "awsSecretAccessKey=${AWS_SECRET_ACCESS_KEY}" \
    --set "awsBillingBucket=${AWS_BILLING_BUCKET}" \
    --set "awsBillingBucketPrefix=${AWS_BILLING_BUCKET_PREFIX}" \
    --set "awsBillingBucketRegion=${AWS_BILLING_BUCKET_REGION}" \
    --set "meteringPullSecretName=${METERING_PULL_SECRET_NAME}" \
    --set "terminationGracePeriodSeconds=${TERMINATION_GRACE_PERIOD_SECONDS}" \
    --set "hdfsNamenodeStorageSize=${HDFS_NAMENODE_STORAGE_SIZE}" \
    --set "hdfsNamenodeMemory=${HDFS_NAMENODE_MEMORY}" \
    --set "hdfsDatanodeStorageSize=${HDFS_DATANODE_STORAGE_SIZE}" \
    --set "hdfsDatanodeMemory=${HDFS_DATANODE_MEMORY}" \
    --set "hiveMetastoreStorageSize=${HIVE_METASTORE_STORAGE_SIZE}" \
    --set "hiveMetastoreMemory=${HIVE_METASTORE_MEMORY}" \
    --set "hiveMetastoreCpu=${HIVE_METASTORE_CPU}" \
    --set "dateAnnotationValue=currdate-${CUR_DATE}" \
)

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    HELM_ARGS+=(--set "imagePullSecretName=$METERING_PULL_SECRET_NAME")
fi

helm template \
    "$ROOT_DIR/charts/metering-ci" \
    -x templates/metering.yaml \
    "${HELM_ARGS[@]}" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$CUSTOM_METERING_CR_FILE"

# use the CUSTOM_METERING_CR_FILE as the CR values for the helm-operator chart values below
CR_SPEC=$("$FAQ_BIN" -f yaml -o yaml -M -c -r '{ cr: {spec: .spec} }' "$CUSTOM_METERING_CR_FILE" )

cat <<EOF > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
image:
  tag: ${DEPLOY_TAG}
annotations: { "metering.deploy-custom/deploy-time": "${CUR_DATE}" }
reconcileIntervalSeconds: 5
${CR_SPEC}
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

"$DIR/deploy-custom.sh"

#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")
ROOT_DIR="$DIR/.."
source "${ROOT_DIR}/hack/common.sh"

TMP_DIR="$(mktemp -d)"

unset METERING_CR_FILE
export CUSTOM_METERING_CR_FILE="$TMP_DIR/custom-metering-cr.yaml"

export METERING_PULL_SECRET_NAME
export METERING_CREATE_PULL_SECRET

: "${METERING_OPERATOR_IMAGE_REPO:=""}"
: "${METERING_OPERATOR_IMAGE_TAG:=""}"
: "${REPORTING_OPERATOR_IMAGE_REPO:=""}"
: "${REPORTING_OPERATOR_IMAGE_TAG:=""}"

: "${USE_KUBE_114_metrics:=true}"
: "${ENABLE_AWS_BILLING:=false}"
: "${DISABLE_PROMETHEUS_METRICS_IMPORTER:=false}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"
: "${AWS_BILLING_BUCKET_REGION:=}"
: "${METERING_CREATE_PULL_SECRET:=false}"
: "${METERING_PULL_SECRET_NAME:=metering-pull-secret}"
: "${TERMINATION_GRACE_PERIOD_SECONDS:=0}"
: "${REPORTING_OPERATOR_REPLICAS:=1}"
: "${REPORTING_OPERATOR_MEMORY:=250Mi}"
: "${REPORTING_OPERATOR_CPU:=1}"
: "${HDFS_NAMENODE_STORAGE_SIZE:=5Gi}"
: "${HDFS_NAMENODE_MEMORY:=500Mi}"
: "${HDFS_DATANODE_STORAGE_SIZE:=5Gi}"
: "${HDFS_DATANODE_MEMORY:=500Mi}"
: "${HIVE_METASTORE_STORAGE_SIZE:=5Gi}"
: "${HIVE_METASTORE_MEMORY:=650Mi}"
: "${HIVE_METASTORE_CPU:=1}"
: "${HIVE_SERVER_MEMORY:=650Mi}"
: "${HIVE_SERVER_CPU:=500m}"
: "${PRESTO_MEMORY:=1Gi}"
: "${PRESTO_CPU:=1}"
: "${CUR_DATE:=$(date +%s)}"

if [ "$DEPLOY_REPORTING_OPERATOR_LOCAL" == "true" ]; then
    REPORTING_OPERATOR_REPLICAS=0
***REMOVED***

HELM_ARGS=(\
    --set "reportingOperatorReplicas=${REPORTING_OPERATOR_REPLICAS}" \
    --set "useKube114Metrics=${USE_KUBE_114_metrics}" \
    --set "enableAwsBilling=${ENABLE_AWS_BILLING}" \
    --set "disablePrometheusMetricsImporter=${DISABLE_PROMETHEUS_METRICS_IMPORTER}" \
    --set "awsAccessKeyId=${AWS_ACCESS_KEY_ID}" \
    --set "awsSecretAccessKey=${AWS_SECRET_ACCESS_KEY}" \
    --set "awsBillingBucket=${AWS_BILLING_BUCKET}" \
    --set "awsBillingBucketPre***REMOVED***x=${AWS_BILLING_BUCKET_PREFIX}" \
    --set "awsBillingBucketRegion=${AWS_BILLING_BUCKET_REGION}" \
    --set "meteringPullSecretName=${METERING_PULL_SECRET_NAME}" \
    --set "terminationGracePeriodSeconds=${TERMINATION_GRACE_PERIOD_SECONDS}" \
    --set "reportingOperatorMemory=${REPORTING_OPERATOR_MEMORY}" \
    --set "reportingOperatorCpu=${REPORTING_OPERATOR_CPU}" \
    --set "hdfsNamenodeStorageSize=${HDFS_NAMENODE_STORAGE_SIZE}" \
    --set "hdfsNamenodeMemory=${HDFS_NAMENODE_MEMORY}" \
    --set "hdfsDatanodeStorageSize=${HDFS_DATANODE_STORAGE_SIZE}" \
    --set "hdfsDatanodeMemory=${HDFS_DATANODE_MEMORY}" \
    --set "hiveMetastoreStorageSize=${HIVE_METASTORE_STORAGE_SIZE}" \
    --set "hiveMetastoreMemory=${HIVE_METASTORE_MEMORY}" \
    --set "hiveMetastoreCpu=${HIVE_METASTORE_CPU}" \
    --set "hiveServerMemory=${HIVE_SERVER_MEMORY}" \
    --set "hiveServerCpu=${HIVE_SERVER_CPU}" \
    --set "prestoMemory=${PRESTO_MEMORY}" \
    --set "prestoCpu=${PRESTO_CPU}" \
    --set "dateAnnotationValue=currdate-${CUR_DATE}" \
)

if [ -n "${REPORTING_OPERATOR_IMAGE_REPO}" ]; then
    HELM_ARGS+=(--set "reportingOperatorDeployRepo=${REPORTING_OPERATOR_IMAGE_REPO}")
***REMOVED***
if [ -n "${REPORTING_OPERATOR_IMAGE_TAG}" ]; then
    HELM_ARGS+=(--set "reportingOperatorDeployTag=${REPORTING_OPERATOR_IMAGE_TAG}")
***REMOVED***

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    HELM_ARGS+=(--set "imagePullSecretName=$METERING_PULL_SECRET_NAME")
***REMOVED***

export METERING_CR_FILE=$CUSTOM_METERING_CR_FILE

helm template \
    "$ROOT_DIR/charts/metering-ci" \
    -x templates/metering.yaml \
    "${HELM_ARGS[@]}" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$CUSTOM_METERING_CR_FILE"

"$DIR/deploy-custom.sh"

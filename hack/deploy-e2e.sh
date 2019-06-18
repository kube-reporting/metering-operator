#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")
ROOT_DIR="$DIR/.."
source "${ROOT_DIR}/hack/common.sh"

: "${METERING_OPERATOR_DEPLOY_REPO:?}"
: "${REPORTING_OPERATOR_DEPLOY_REPO:?}"
: "${METERING_OPERATOR_DEPLOY_TAG:?}"
: "${REPORTING_OPERATOR_DEPLOY_TAG:?}"

TMP_DIR="$(mktemp -d)"

unset METERING_CR_FILE
export CUSTOM_METERING_CR_FILE="$TMP_DIR/custom-metering-cr.yaml"
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-ansible-operator-values.yaml"}
export CUSTOM_OLM_OVERRIDE_VALUES=${CUSTOM_OLM_OVERRIDE_VALUES:-"$TMP_DIR/custom-olm-values.yaml"}

export DEPLOY_PLATFORM
export METERING_PULL_SECRET_NAME
export METERING_CREATE_PULL_SECRET

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
fi

HELM_ARGS=(\
    --set "reportingOperatorReplicas=${REPORTING_OPERATOR_REPLICAS}" \
    --set "reportingOperatorDeployRepo=${REPORTING_OPERATOR_DEPLOY_REPO}" \
    --set "reportingOperatorDeployTag=${REPORTING_OPERATOR_DEPLOY_TAG}" \
    --set "enableAwsBilling=${ENABLE_AWS_BILLING}" \
    --set "disablePrometheusMetricsImporter=${DISABLE_PROMETHEUS_METRICS_IMPORTER}" \
    --set "awsAccessKeyId=${AWS_ACCESS_KEY_ID}" \
    --set "awsSecretAccessKey=${AWS_SECRET_ACCESS_KEY}" \
    --set "awsBillingBucket=${AWS_BILLING_BUCKET}" \
    --set "awsBillingBucketPrefix=${AWS_BILLING_BUCKET_PREFIX}" \
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

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    HELM_ARGS+=(--set "imagePullSecretName=$METERING_PULL_SECRET_NAME")
fi

export METERING_CR_FILE=$CUSTOM_METERING_CR_FILE
helm template \
    "$ROOT_DIR/charts/metering-ci" \
    -x templates/metering.yaml \
    "${HELM_ARGS[@]}" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$CUSTOM_METERING_CR_FILE"

cat <<EOF > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
operator:
  image:
    repo: ${METERING_OPERATOR_DEPLOY_REPO}
    tag: ${METERING_OPERATOR_DEPLOY_TAG}
  annotations: { "metering.deploy-custom/deploy-time": "${CUR_DATE}" }
  reconcileIntervalSeconds: 5
EOF

"$DIR/deploy-custom.sh"

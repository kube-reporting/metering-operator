#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${DEPLOY_TAG:?}"

# Used in deploy.sh
export DOCKER_USERNAME="${DOCKER_CREDS_USR:-}"
export DOCKER_PASSWORD="${DOCKER_CREDS_PSW:-}"

export METERING_PULL_SECRET_NAME
export METERING_CREATE_PULL_SECRET
export UNINSTALL_METERING_BEFORE_INSTALL="${UNINSTALL_METERING_BEFORE_INSTALL:-false}"
# Do not uninstall metering after deploying
export CLEANUP_METERING=false

export DISABLE_PROMSUM=false
export HDFS_NAMENODE_STORAGE_SIZE="20Gi"
export HDFS_NAMENODE_MEMORY="650Mi"
export HDFS_DATANODE_STORAGE_SIZE="30Gi"
export HDFS_DATANODE_MEMORY="800Mi"
export HIVE_METASTORE_MEMORY="1500Mi"
export HIVE_METASTORE_CPU="1"
export HIVE_METASTORE_STORAGE_SIZE="20Gi"
export TERMINATION_GRACE_PERIOD_SECONDS=60

"$DIR/deploy-e2e.sh"

echo "Deploying default Reports"

HOURLY=( \
    "$MANIFESTS_DIR/reports/cluster-capacity-hourly.yaml" \
    "$MANIFESTS_DIR/reports/cluster-usage-hourly.yaml" \
    "$MANIFESTS_DIR/reports/cluster-utilization-hourly.yaml" \
    "$MANIFESTS_DIR/reports/namespace-usage-hourly.yaml" \
)
DAILY=( \
    "$MANIFESTS_DIR/reports/cluster-capacity-daily.yaml" \
    "$MANIFESTS_DIR/reports/cluster-usage-daily.yaml" \
    "$MANIFESTS_DIR/reports/cluster-utilization-daily.yaml" \
    "$MANIFESTS_DIR/reports/namespace-usage-daily.yaml" \
)

echo "Creating hourly reports"
kube-install "${HOURLY[@]}"
echo "Creating daily reports"
kube-install "${DAILY[@]}"

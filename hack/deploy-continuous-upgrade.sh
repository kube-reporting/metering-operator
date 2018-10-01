#!/bin/bash
set -e

DIR=$(dirname "${BASH_SOURCE}")

# Used in deploy.sh
export DOCKER_USERNAME="$DOCKER_CREDS_USR"
export DOCKER_PASSWORD="$DOCKER_CREDS_PSW"

export HDFS_NAMENODE_STORAGE_SIZE="20Gi"
export HDFS_DATANODE_STORAGE_SIZE="30Gi"

export METERING_CREATE_PULL_SECRET=true
"$DIR/deploy-custom.sh"

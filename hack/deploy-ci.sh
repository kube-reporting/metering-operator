#!/bin/bash
set -e

: "${DEPLOY_TAG:?}"
: "${DEPLOY_PLATFORM:?must be set to either tectonic, openshift, or generic}"

TMP_DIR="$(mktemp -d)"

export INSTALL_METHOD="${DEPLOY_PLATFORM}-direct"
export METERING_CR_FILE=${METERING_CR_FILE:-"$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"}
export CUSTOM_DEPLOY_MANIFESTS_DIR=${CUSTOM_DEPLOY_MANIFESTS_DIR:-"$TMP_DIR/custom-deploy-manifests-${DEPLOY_TAG}"}
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}
export DELETE_PVCS=${DELETE_PVCS:-true}

# Used in deploy.sh
export DOCKER_USERNAME="$DOCKER_CREDS_USR"
export DOCKER_PASSWORD="$DOCKER_CREDS_PSW"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${ENABLE_AWS_BILLING:=false}"
: "${DISABLE_PROMSUM:=true}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"
: "${METERING_CREATE_PULL_SECRET:=true}"
: "${METERING_PULL_SECRET_NAME:=metering-pull-secret}"
: "${TERMINATION_GRACE_PERIOD_SECONDS:=}"

IMAGE_PULL_SECRET_TEXT=""
if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    IMAGE_PULL_SECRET_TEXT="imagePullSecrets: [ { name: \"$METERING_PULL_SECRET_NAME\" } ]"
fi

cat <<EOF > "$METERING_CR_FILE"
apiVersion: chargeback.coreos.com/v1alpha1
kind: Metering
metadata:
  name: "${DEPLOY_PLATFORM}-metering"
spec:
  metering-operator:
    image:
      tag: ${DEPLOY_TAG}

    ${IMAGE_PULL_SECRET_TEXT:-}
    terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}

    config:
      disablePromsum: ${DISABLE_PROMSUM}
      awsBillingDataSource:
        enabled: ${ENABLE_AWS_BILLING}
        bucket: "${AWS_BILLING_BUCKET}"
        prefix: "${AWS_BILLING_BUCKET_PREFIX}"
      awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
      awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"


  presto:
    ${IMAGE_PULL_SECRET_TEXT:-}
    config:
      awsAccessKeyID: "${AWS_ACCESS_KEY_ID}"
      awsSecretAccessKey: "${AWS_SECRET_ACCESS_KEY}"
    presto:
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
      image:
        tag: ${DEPLOY_TAG}
    hive:
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
      image:
        tag: ${DEPLOY_TAG}

  hdfs:
    image:
      tag: ${DEPLOY_TAG}
    ${IMAGE_PULL_SECRET_TEXT:-}
    datanode:
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
    namenode:
      terminationGracePeriodSeconds: ${TERMINATION_GRACE_PERIOD_SECONDS}
EOF



cat <<EOF > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
image:
  tag: ${DEPLOY_TAG}
reconcileIntervalSeconds: 5
${IMAGE_PULL_SECRET_TEXT:-}
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
export MANIFEST_OUTPUT_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"$ROOT_DIR/hack/create-metering-manifests.sh"

echo "Deploying"
export DEPLOY_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"${ROOT_DIR}/hack/deploy.sh"

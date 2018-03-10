#!/bin/bash
set -e

: ${DEPLOY_TAG:?}

TMP_DIR="$(mktemp -d)"
export CHARGEBACK_CR_FILE=${CHARGEBACK_CR_FILE:-"$TMP_DIR/custom-chargeback-cr-${DEPLOY_TAG}.yaml"}
export INSTALLER_MANIFEST_DIR=${INSTALLER_MANIFEST_DIR:-"$TMP_DIR/installer_manifests-${DEPLOY_TAG}"}
export CUSTOM_VALUES_FILE=${CUSTOM_VALUES_FILE:-"$TMP_DIR/helm-operator-values-${DEPLOY_TAG}.yaml"}
export DELETE_PVCS=${DELETE_PVCS:-true}

: "${ENABLE_AWS_BILLING:=false}"
: "${AWS_ACCESS_KEY_ID:=}"
: "${AWS_SECRET_ACCESS_KEY:=}"
: "${AWS_BILLING_BUCKET:=}"
: "${AWS_BILLING_BUCKET_PREFIX:=}"

cat <<EOF > "$CHARGEBACK_CR_FILE"
apiVersion: chargeback.coreos.com/v1alpha1
kind: Chargeback
metadata:
  name: "openshift-chargeback"
spec:
  chargeback-operator:
    image:
      tag: ${DEPLOY_TAG}

    con***REMOVED***g:
      disablePromsum: true
      prometheusURL: "http://prometheus-k8s.monitoring.svc.cluster.local:9090/"

  presto:
    presto:
      terminationGracePeriodSeconds: 0
      securityContext:
        fsGroup: null
      image:
        tag: ${DEPLOY_TAG}
    hive:
      securityContext:
        fsGroup: null
      terminationGracePeriodSeconds: 0
      image:
        tag: ${DEPLOY_TAG}

  hdfs:
    image:
      tag: ${DEPLOY_TAG}
    securityContext:
      fsGroup: null
    datanode:
      terminationGracePeriodSeconds: 0
    namenode:
      terminationGracePeriodSeconds: 0
EOF


cat <<EOF > "$CUSTOM_VALUES_FILE"
name: chargeback-helm-operator
image:
  tag: ${DEPLOY_TAG}
reconcileIntervalSeconds: 5
EOF

echo "Creating installer manifests"

./hack/create-installer-manifests.sh "$CUSTOM_VALUES_FILE"

echo "Deploying"

./hack/deploy.sh

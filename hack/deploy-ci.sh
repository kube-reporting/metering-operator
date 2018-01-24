#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
: ${DEPLOY_TAG:?}

export CHARGEBACK_CR_FILE="/tmp/custom-chargeback-cr-${DEPLOY_TAG}.yaml"
export INSTALLER_MANIFEST_DIR="/tmp/installer_manifests-${DEPLOY_TAG}"
export DELETE_PVCS=true

cat <<EOF > "$CHARGEBACK_CR_FILE"
apiVersion: chargeback.coreos.com/v1alpha1
kind: Chargeback
metadata:
  name: "tectonic-chargeback"
spec:
  chargeback-operator:
    image:
      tag: ${DEPLOY_TAG}

      config:
        disablePromsum: true

  presto:
    presto:
      image:
        tag: ${DEPLOY_TAG}
    hive:
      image:
        tag: ${DEPLOY_TAG}

  hdfs:
    image:
      tag: ${DEPLOY_TAG}
EOF

CUSTOM_VALUES_FILE="/tmp/helm-operator-values-${DEPLOY_TAG}.yaml"

cat <<EOF > "$CUSTOM_VALUES_FILE"
name: chargeback-helm-operator
operator:
  image:
    tag: ${DEPLOY_TAG}
EOF

./hack/create-installer-manifests.sh \
    "$DIR/chargeback-helm-operator-values.yaml" \
    "$CUSTOM_VALUES_FILE"

./hack/deploy.sh

#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
: ${DEPLOY_TAG:?}

export CHARGEBACK_CR_FILE="/tmp/custom-chargeback-cr-${DEPLOY_TAG}.yaml"

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

./hack/deploy.sh

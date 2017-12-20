#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
: ${DEPLOY_TAG:?}

export CUSTOM_CHARGEBACK_SETTINGS_FILE="/tmp/custom-values-${DEPLOY_TAG}.yaml"

cat <<EOF > "$CUSTOM_CHARGEBACK_SETTINGS_FILE"
chargeback-operator:
  image:
    tag: ${DEPLOY_TAG}

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

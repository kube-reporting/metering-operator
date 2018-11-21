#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
fi

if [ "$METERING_INSTALL_REPORTING_OPERATOR_EXTRA_CLUSTERROLEBINDING" == "true" ]; then
    echo "Creating ClusterRole for reporting-operator"
    kubectl \
        apply -f \
        "${DEPLOY_MANIFESTS_DIR}/reporting-operator-clusterrole.yaml"

    echo "Creating ClusterRoleBinding for reporting-operator"
    kubectl \
        create clusterrolebinding \
        "${METERING_REPORTING_OPERATOR_EXTRA_ROLEBINDING_NAME}" \
        --clusterrole \
        "${METERING_REPORTING_OPERATOR_EXTRA_ROLE_NAME}" \
        --serviceaccount \
        "${METERING_NAMESPACE}:reporting-operator" \
        --dry-run -o json | kubectl replace -f -
fi

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/install.sh" "$@"

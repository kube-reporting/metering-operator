set -x
set -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..

REPO_NAMESPACE="${REPO_NAMESPACE:-"openshift"}"
OSE_IMAGE_TAG="${OSE_IMAGE_TAG:-"v4.3"}"
CLUSTER_REGISTRY_URL="${CLUSTER_REGISTRY_URL:-"$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')"}"
SETUP_REGISTRY_AUTH="${SETUP_REGISTRY_AUTH:-"true"}"
PULL_IMAGES="${PULL_IMAGES:-"true"}"
PUSH_IMAGES="${PUSH_IMAGES:-"true"}"

if [ -z "$CLUSTER_REGISTRY_URL" ]; then
    echo "Couldn't detect \$CLUSTER_REGISTRY_URL or unset"
    exit 1
fi

# default to mirroring the $OSE_IMAGE_TAG for each, but allow overriding the tag to be mirrored per mirror
METERING_ANSIBLE_OPERATOR_IMAGE_TAG="${METERING_ANSIBLE_OPERATOR_IMAGE_TAG:-$OSE_IMAGE_TAG}"
METERING_REPORTING_OPERATOR_IMAGE_TAG="${METERING_REPORTING_OPERATOR_IMAGE_TAG:-$OSE_IMAGE_TAG}"
METERING_PRESTO_IMAGE_TAG="${METERING_PRESTO_IMAGE_TAG:-$OSE_IMAGE_TAG}"
METERING_HIVE_IMAGE_TAG="${METERING_HIVE_IMAGE_TAG:-$OSE_IMAGE_TAG}"
METERING_HADOOP_IMAGE_TAG="${METERING_HADOOP_IMAGE_TAG:-$OSE_IMAGE_TAG}"
GHOSTUNNEL_IMAGE_TAG="${GHOSTUNNEL_IMAGE_TAG:-$OSE_IMAGE_TAG}"
OAUTH_PROXY_IMAGE_TAG="${OAUTH_PROXY_IMAGE_TAG:-$OSE_IMAGE_TAG}"

METERING_ANSIBLE_OPERATOR_IMAGE="${METERING_ANSIBLE_OPERATOR_IMAGE:-"openshift/ose-metering-ansible-operator:$METERING_ANSIBLE_OPERATOR_IMAGE_TAG"}"
METERING_REPORTING_OPERATOR_IMAGE="${METERING_REPORTING_OPERATOR_IMAGE:-"openshift/ose-metering-reporting-operator:$METERING_REPORTING_OPERATOR_IMAGE_TAG"}"
METERING_PRESTO_IMAGE="${METERING_PRESTO_IMAGE:-"openshift/ose-metering-presto:$METERING_PRESTO_IMAGE_TAG"}"
METERING_HIVE_IMAGE="${METERING_HIVE_IMAGE:-"openshift/ose-metering-hive:$METERING_HIVE_IMAGE_TAG"}"
METERING_HADOOP_IMAGE="${METERING_HADOOP_IMAGE:-"openshift/ose-metering-hadoop:$METERING_HADOOP_IMAGE_TAG"}"
GHOSTUNNEL_IMAGE="${GHOSTUNNEL_IMAGE:-"openshift/ose-ghostunnel:$GHOSTUNNEL_IMAGE_TAG"}"
OAUTH_PROXY_IMAGE="${OAUTH_PROXY_IMAGE:-"openshift/ose-oauth-proxy:$OAUTH_PROXY_IMAGE_TAG"}"

: "${METERING_NAMESPACE:?"\$METERING_NAMESPACE must be set!"}"

if [ "$SETUP_REGISTRY_AUTH" == "true" ]; then
    echo "Creating namespace for images: $REPO_NAMESPACE"
    oc create namespace "$REPO_NAMESPACE" || true
    echo "Creating serviceaccount registry-editor in $REPO_NAMESPACE"
    oc create serviceaccount registry-editor -n "$REPO_NAMESPACE" || true
    echo "Granting registry-editor registry-editor permissions in $REPO_NAMESPACE"
    oc adm policy add-role-to-user registry-editor -z registry-editor -n "$REPO_NAMESPACE" || true
    echo "Performing docker login as registry-editor to $CLUSTER_REGISTRY_URL"
    set +x
    docker login \
        "$CLUSTER_REGISTRY_URL" \
        -u registry-editor \
        -p "$(oc sa get-token registry-editor -n "$REPO_NAMESPACE")"
    set -x
fi

echo "Ensuring namespace $REPO_NAMESPACE exists for images to be pushed into"
oc create namespace "$REPO_NAMESPACE" || true
echo "Pushing Metering OSE images to $CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$METERING_ANSIBLE_OPERATOR_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$METERING_REPORTING_OPERATOR_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$METERING_PRESTO_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$METERING_HIVE_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$METERING_HADOOP_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$GHOSTUNNEL_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

"$ROOT_DIR/hack/mirror-ose-image.sh" \
    "$OAUTH_PROXY_IMAGE" \
    "$CLUSTER_REGISTRY_URL"

echo "Granting access to pull images in $REPO_NAMESPACE to all serviceaccounts in \$METERING_NAMESPACE=$METERING_NAMESPACE"
oc -n "$REPO_NAMESPACE" policy add-role-to-group system:image-puller "system:serviceaccounts:$METERING_NAMESPACE" --rolebinding-name "$METERING_NAMESPACE-image-pullers"

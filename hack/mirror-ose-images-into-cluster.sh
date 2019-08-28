set -x

REPO_NAMESPACE="${REPO_NAMESPACE:-"openshift-metering-images"}"
OSE_IMAGE_TAG="${OSE_IMAGE_TAG:-"v4.2"}"
OSE_IMAGE_REPO="${OSE_IMAGE_REPO:-"brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888"}"
CLUSTER_REGISTRY_URL="${CLUSTER_REGISTRY_URL:-"$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')"}"
SETUP_REGISTRY_AUTH="${SETUP_REGISTRY_AUTH:-"true"}"
PULL_IMAGES="${PULL_IMAGES:-"true"}"
PUSH_IMAGES="${PUSH_IMAGES:-"true"}"

: "${METERING_NAMESPACE:?"\$METERING_NAMESPACE must be set!"}"

if [ "$SETUP_REGISTRY_AUTH" == "true" ]; then
    echo "Creating namespace for images: $REPO_NAMESPACE"
    oc create namespace "$REPO_NAMESPACE" || true
    echo "Creating serviceaccount registry-editor in $REPO_NAMESPACE"
    oc create serviceaccount registry-editor -n "$REPO_NAMESPACE"
    echo "Granting registry-editor registry-editor permissions in $REPO_NAMESPACE"
    oc adm policy add-role-to-user registry-editor -z registry-editor -n "$REPO_NAMESPACE"
    echo "Performing docker login as registry-editor to $CLUSTER_REGISTRY_URL"
    set +x
    docker login \
        "$CLUSTER_REGISTRY_URL" \
        -u registry-editor \
        -p "$(oc sa get-token registry-editor -n "$REPO_NAMESPACE")"
    set -x
fi

if [ "$PULL_IMAGES" == "true" ]; then
    echo "Pulling Metering OSE images from $OSE_IMAGE_REPO"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-metering-ansible-operator:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-metering-reporting-operator:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-metering-presto:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-metering-hive:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-metering-hadoop:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-ghostunnel:$OSE_IMAGE_TAG"
    docker pull "$OSE_IMAGE_REPO/openshift/ose-oauth-proxy:$OSE_IMAGE_TAG"
fi

if [ "$PUSH_IMAGES" == "true" ]; then
    echo "Ensuring namespace $REPO_NAMESPACE exists for images to be pushed into"
    oc create namespace "$REPO_NAMESPACE" || true
    echo "Pushing Metering OSE images to $CLUSTER_REGISTRY_URL"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-metering-ansible-operator:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-ansible-operator:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-ansible-operator:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-metering-reporting-operator:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-reporting-operator:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-reporting-operator:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-metering-presto:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-presto:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-presto:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-metering-hive:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-hive:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-hive:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-metering-hadoop:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-hadoop:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-metering-hadoop:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-ghostunnel:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-ghostunnel:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-ghostunnel:$OSE_IMAGE_TAG"

    docker tag \
        "$OSE_IMAGE_REPO/openshift/ose-oauth-proxy:$OSE_IMAGE_TAG" \
        "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-oauth-proxy:$OSE_IMAGE_TAG"
    docker push "$CLUSTER_REGISTRY_URL/$REPO_NAMESPACE/ose-oauth-proxy:$OSE_IMAGE_TAG"

    echo "Granting access to pull images in $REPO_NAMESPACE to all serviceaccounts in \$METERING_NAMESPACE=$METERING_NAMESPACE"
    oc -n openshift-metering-images policy add-role-to-group system:image-puller "system:serviceaccounts:$METERING_NAMESPACE" --rolebinding-name "$METERING_NAMESPACE-image-pullers"
fi

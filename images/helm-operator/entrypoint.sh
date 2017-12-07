#!/bin/bash

set -e

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
***REMOVED***

: ${HELM_CHART_PATH:?}
: ${HELM_RELEASE_NAME:?}
: ${HELM_VALUES_SECRET_NAME:?}
: ${HELM_WAIT:=false}
: ${HELM_WAIT_TIMEOUT:=120}

: ${MY_POD_NAMESPACE:?}

: ${HELM_RECONCILE_INTERVAL_SECONDS:=120}
: ${HELM_HOST:="127.0.0.1:44134"}

: ${TILLER_READY_ENDPOINT:="127.0.0.1:44135/readiness"}

export HELM_HOST

NEEDS_EXIT=false

trap setNeedsExit SIGINT SIGTERM

setNeedsExit() {
    echo "Got shutdown signal"
    NEEDS_EXIT=true
}

checkExit() {
    if [ "$NEEDS_EXIT" == "true" ]; then
        echo "***REMOVED***nished shutdown, exiting"
        exit 0
    ***REMOVED***
}

until curl -s $TILLER_READY_ENDPOINT; do
    echo "Waiting for Tiller to become ready"
    sleep 1
done

EXTRA_ARGS=()
if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
    echo "Getting pod $MY_POD_NAME owner information"
    source get_owner.sh

    cat <<EOF > /tmp/owner-values.yaml
global:
  ownerReferences:
  - apiVersion: "apps/v1beta1"
    blockOwnerDeletion: false
    controller: true
    kind: "Deployment"
    name: $MY_DEPLOYMENT_NAME
    uid: $MY_DEPLOYMENT_UID
EOF


    cat <<EOF > /tmp/owner-patch.json
{
  "metadata": {
    "ownerReferences": [{
      "apiVersion": "apps/v1beta1",
      "blockOwnerDeletion": false,
      "controller": true,
      "kind": "Deployment",
      "name": "$MY_DEPLOYMENT_NAME",
      "uid": "$MY_DEPLOYMENT_UID"
    }]
  }
}
EOF

    echo "Owner references: "
    echo "$(cat /tmp/owner-values.yaml)"
    EXTRA_ARGS+=(-f /tmp/owner-values.yaml)
***REMOVED***

while true; do
    checkExit

    echo "Fetching helm values from secret $HELM_VALUES_SECRET_NAME"
    touch /tmp/values.yaml
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get secrets "$HELM_VALUES_SECRET_NAME" \
        --ignore-not-found \
        -o json > "${HELM_VALUES_SECRET_NAME}.json"

    if [ -s "${HELM_VALUES_SECRET_NAME}.json" ]; then
        echo "Got secret ${HELM_VALUES_SECRET_NAME}"
        jq '.data["values.yaml"]' ${HELM_VALUES_SECRET_NAME}.json -r > /tmp/values.json
        if [ "$(cat /tmp/values.json)" != "null" ]; then
            base64 -d /tmp/values.json > /tmp/values.yaml
        ***REMOVED***
            echo "No values.yaml found in ${HELM_VALUES_SECRET_NAME}"
        ***REMOVED***
        rm -f /tmp/values.json
    ***REMOVED***
        echo "Secret ${HELM_VALUES_SECRET_NAME} does not exist, default values will be used"
    ***REMOVED***

    echo "Running helm upgrade"
    set +e
    helm upgrade \
        --install \
        --namespace "$MY_POD_NAMESPACE" \
        --wait="$HELM_WAIT" \
        --timeout="$HELM_WAIT_TIMEOUT" \
        "$HELM_RELEASE_NAME"\
        "$HELM_CHART_PATH" \
        -f "/tmp/values.yaml" \
        "${EXTRA_ARGS[@]}" "$@"
    HELM_EXIT_CODE=$?
    if [ $HELM_EXIT_CODE != 0 ]; then
        echo "helm upgrade failed, exit code: $HELM_EXIT_CODE"
    ***REMOVED***
    set -e

    RELEASE_CMS=$(kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get con***REMOVED***gmap \
        -l "OWNER=TILLER,NAME=$HELM_RELEASE_NAME" \
        -o json | jq '.' -cr)

    if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
        echo "Setting ownerReferences for Helm release con***REMOVED***gmaps"

        RELEASE_CM_NAMES=$(echo $RELEASE_CMS | jq '.items[] | select(.metadata.ownerReferences | length == 0) | .metadata.name' -r)
        for cm in $RELEASE_CM_NAMES; do
            kubectl \
                --namespace "$MY_POD_NAMESPACE" \
                patch con***REMOVED***gmap $cm \
                -p "$(cat /tmp/owner-patch.json)"

        done
    ***REMOVED***

    if [ -n "$RELEASE_HISTORY_LIMIT" ]; then
        echo "Getting list of helm release con***REMOVED***gmaps to delete"
        DELETE_RELEASE_CM_NAMES=$(echo $RELEASE_CMS | jq '.items | length as $listLength | ($listLength - (env.RELEASE_HISTORY_LIMIT | tonumber)) as $limitSize | (if $limitSize < 0 then 0 ***REMOVED*** $limitSize end) as $limitSize | sort_by(.metadata.labels.VERSION | tonumber) | limit($limitSize; .[]) | .metadata.name' -r)
        if [ -z "$DELETE_RELEASE_CM_NAMES" ]; then
            echo "No release con***REMOVED***gmaps to delete yet"
        ***REMOVED***
            for cm in $DELETE_RELEASE_CM_NAMES; do
                    echo "Deleting helm release con***REMOVED***gmap $cm"
                    kubectl \
                        --namespace "$MY_POD_NAMESPACE" \
                        delete con***REMOVED***gmap $cm
            done
        ***REMOVED***
    ***REMOVED***

    checkExit

    echo "Sleeping $HELM_RECONCILE_INTERVAL_SECONDS seconds"
    sleep $HELM_RECONCILE_INTERVAL_SECONDS
done

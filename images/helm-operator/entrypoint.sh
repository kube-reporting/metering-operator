#!/bin/bash

set -e

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
***REMOVED***

: ${HELM_CHART_PATH:?}
: ${HELM_RELEASE_CRD_NAME:?}
: ${HELM_RELEASE_CRD_API_GROUP:?}

: ${HELM_WAIT:=false}
: ${HELM_WAIT_TIMEOUT:=120}

: ${MY_POD_NAMESPACE:?}

: ${HELM_RECONCILE_INTERVAL_SECONDS:=120}
: ${HELM_HOST:="127.0.0.1:44134"}

: ${TILLER_READY_ENDPOINT:="127.0.0.1:44135/readiness"}

export HELM_HOST
export RELEASE_HISTORY_LIMIT

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

getReleaseCon***REMOVED***gmaps() {
    HELM_RELEASE_NAME=$1
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get con***REMOVED***gmap \
        -l "OWNER=TILLER,NAME=$HELM_RELEASE_NAME" \
        -o json | jq '.' -r
}

setOwnerOnReleaseCon***REMOVED***gmaps(){
    if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
        echo "Setting ownerReferences for Helm release con***REMOVED***gmaps"

        RELEASE_CM_NAMES=$(jq '.items[] | select(.metadata.ownerReferences | length == 0) | .metadata.name' -r $1)
        if [ -z "$RELEASE_CM_NAMES" ]; then
            echo "No release con***REMOVED***gmaps to patch ownership of yet"
        ***REMOVED***
            echo -n "$RELEASE_CM_NAMES" | while read -r cm; do
                echo "Setting owner of $cm to deployment $MY_DEPLOYMENT_NAME - $MY_DEPLOYMENT_UID"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    patch con***REMOVED***gmap $cm \
                    -p "$(cat /tmp/owner-patch.json)"
            done
        ***REMOVED***
    ***REMOVED***
}

cleanupOldReleaseCon***REMOVED***gmaps() {
    if [ -n "$RELEASE_HISTORY_LIMIT" ]; then
        echo "Getting list of helm release con***REMOVED***gmaps to delete"
        DELETE_RELEASE_CM_NAMES=$(jq '.items | length as $listLength | ($listLength - (env.RELEASE_HISTORY_LIMIT | tonumber)) as $limitSize | (if $limitSize < 0 then 0 ***REMOVED*** $limitSize end) as $limitSize | sort_by(.metadata.labels.VERSION | tonumber) | limit($limitSize; .[]) | .metadata.name' -rc $1)
        if [ -z "$DELETE_RELEASE_CM_NAMES" ]; then
            echo "No release con***REMOVED***gmaps to delete yet"
        ***REMOVED***
            echo -n "$DELETE_RELEASE_CM_NAMES" | while read -r cm; do
                echo "Deleting helm release con***REMOVED***gmap $cm"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    delete con***REMOVED***gmap $cm
            done
        ***REMOVED***
    ***REMOVED***
}

writeReleaseCon***REMOVED***gMapOwnerPatchFile() {
    OWNER_KIND=$1
    OWNER_NAME=$2
    OWNER_UID=$3
    cat <<EOF > /tmp/owner-patch.json
{
  "metadata": {
    "ownerReferences": [{
      "apiVersion": "apps/v1beta1",
      "blockOwnerDeletion": false,
      "controller": true,
      "kind": "$OWNER_KIND",
      "name": "$OWNER_NAME",
      "uid": "$OWNER_UID"
    }]
  }
}
EOF
}

writeReleaseOwnerValuesFile() {
    OWNER_KIND=$1
    OWNER_NAME=$2
    OWNER_UID=$3
    cat <<EOF > /tmp/owner-values.yaml
global:
  ownerReferences:
  - apiVersion: "apps/v1beta1"
    blockOwnerDeletion: false
    controller: true
    kind: "$OWNER_KIND"
    name: "$OWNER_NAME"
    uid: "$OWNER_UID"
EOF
}

helmUpgrade() {
    RELEASE_NAME=$1
    helm upgrade \
        --install \
        --namespace "$MY_POD_NAMESPACE" \
        --wait="$HELM_WAIT" \
        --timeout="$HELM_WAIT_TIMEOUT" \
        "$RELEASE_NAME"\
        "$HELM_CHART_PATH" \
        "${@:2}"
    HELM_EXIT_CODE=$?
    if [ $HELM_EXIT_CODE != 0 ]; then
        echo "helm upgrade failed, exit code: $HELM_EXIT_CODE"
    ***REMOVED***
}

until curl -s $TILLER_READY_ENDPOINT; do
    echo "Waiting for Tiller to become ready"
    sleep 1
done

getReleaseCon***REMOVED***gmaps > /tmp/release-con***REMOVED***gmaps.json
cleanupOldReleaseCon***REMOVED***gmaps /tmp/release-con***REMOVED***gmaps.json
checkExit

if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
    echo "Getting pod $MY_POD_NAME owner information"
    source get_owner.sh
    writeReleaseCon***REMOVED***gMapOwnerPatchFile "Deployment" "$MY_DEPLOYMENT_NAME" "$MY_DEPLOYMENT_UID"

    getReleaseCon***REMOVED***gmaps > /tmp/release-con***REMOVED***gmaps.json
    setOwnerOnReleaseCon***REMOVED***gmaps /tmp/release-con***REMOVED***gmaps.json
    checkExit
***REMOVED***


while true; do
    checkExit

    CRD="${HELM_RELEASE_CRD_NAME}.${HELM_RELEASE_CRD_API_GROUP}"
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get "$CRD" \
        --ignore-not-found \
        -o json > /tmp/helm-releases.json

    if [ -s /tmp/helm-releases.json ]; then
        while read -r release; do
            echo -E "$release" > /tmp/current-release.json
            RELEASE_NAME="$(jq -Mcr '.metadata.name' /tmp/current-release.json)"
            RELEASE_UID="$(jq -Mcr '.metadata.uid' /tmp/current-release.json)"
            RELEASE_VALUES="$(jq -Mcr '.spec.values // empty' /tmp/current-release.json)"

            if [ -z "$RELEASE_VALUES" ]; then
                echo "No values, using default values"
            ***REMOVED***
                VALUES_FILE="/tmp/${RELEASE_NAME}-values.yaml"
                echo -E "$RELEASE_VALUES" > "$VALUES_FILE"

                HELM_ARGS=("-f" "$VALUES_FILE")
            ***REMOVED***

            writeReleaseOwnerValuesFile "$HELM_RELEASE_CRD_NAME" "$RELEASE_NAME" "$RELEASE_UID"
            EXTRA_ARGS=("-f" /tmp/owner-values.yaml)

            echo "Running helm upgrade for release $RELEASE_NAME"
            helmUpgrade "$RELEASE_NAME" "${EXTRA_ARGS[@]}" "${HELM_ARGS[@]}"

            getReleaseCon***REMOVED***gmaps > /tmp/release-con***REMOVED***gmaps.json
            setOwnerOnReleaseCon***REMOVED***gmaps /tmp/release-con***REMOVED***gmaps.json
            cleanupOldReleaseCon***REMOVED***gmaps /tmp/release-con***REMOVED***gmaps.json
            checkExit
        done < <(jq '.items[]' -Mcr /tmp/helm-releases.json)

        echo "Sleeping $HELM_RECONCILE_INTERVAL_SECONDS seconds"
        for ((i=0; i < $HELM_RECONCILE_INTERVAL_SECONDS; i++)); do
            sleep 1
            checkExit
        done
    ***REMOVED***
        echo "No resources with kind $HELM_RELEASE_CRD_NAME and group $HELM_RELEASE_CRD_API_GROUP"
    ***REMOVED***
done

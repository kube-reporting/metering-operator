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
: ${EXTRA_VALUES_FILE:=}

: ${MY_POD_NAMESPACE:?}

: ${HELM_RECONCILE_INTERVAL_SECONDS:=120}
: ${HELM_HOST:="127.0.0.1:44134"}

: ${TILLER_READY_ENDPOINT:="127.0.0.1:44135/readiness"}

export HELM_HOST
export RELEASE_HISTORY_LIMIT

NEEDS_EXIT=false

trap setNeedsExit SIGINT SIGTERM


OWNER_PATCH_FILE=/tmp/owner-patch.json
OWNER_VALUES_FILE=/tmp/owner-values.yaml
HELM_RELEASES_FILE=/tmp/helm-release.json
CURRENT_RELEASE_FILE=/tmp/current-release.json
RELEASE_CONFIGMAPS_FILE=/tmp/release-con***REMOVED***gmaps.json

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

writeReleaseCon***REMOVED***gmapsFile() {
    HELM_RELEASE_NAME=$1
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get con***REMOVED***gmap \
        -l "OWNER=TILLER,NAME=$HELM_RELEASE_NAME" \
        -o json | jq '.' -r > "$RELEASE_CONFIGMAPS_FILE"
}

setOwnerOnReleaseCon***REMOVED***gmaps(){
    if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
        echo "Setting ownerReferences for Helm release con***REMOVED***gmaps"

        RELEASE_CM_NAMES="$(jq '.items[] | select(.metadata.ownerReferences | length == 0) | .metadata.name' -r "$RELEASE_CONFIGMAPS_FILE")"
        if [ -z "$RELEASE_CM_NAMES" ]; then
            echo "No release con***REMOVED***gmaps to patch ownership of yet"
        ***REMOVED***
            echo "$RELEASE_CM_NAMES" | while read -r cm; do
                echo "Setting owner of $cm"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    patch con***REMOVED***gmap "$cm" \
                    -p "$(cat $OWNER_PATCH_FILE)"
            done
        ***REMOVED***
    ***REMOVED***
}

cleanupOldReleaseCon***REMOVED***gmaps() {
    if [ -n "$RELEASE_HISTORY_LIMIT" ]; then
        echo "Getting list of helm release con***REMOVED***gmaps to delete"
        DELETE_RELEASE_CM_NAMES="$(jq '.items | length as $listLength | ($listLength - (env.RELEASE_HISTORY_LIMIT | tonumber)) as $limitSize | (if $limitSize < 0 then 0 ***REMOVED*** $limitSize end) as $limitSize | sort_by(.metadata.labels.VERSION | tonumber) | limit($limitSize; .[]) | .metadata.name' -rc "$RELEASE_CONFIGMAPS_FILE")"
        if [ -z "$DELETE_RELEASE_CM_NAMES" ]; then
            echo "No release con***REMOVED***gmaps to delete yet"
        ***REMOVED***
            echo "$DELETE_RELEASE_CM_NAMES" | while read -r cm; do
                echo "Deleting helm release con***REMOVED***gmap $cm"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    delete con***REMOVED***gmap "$cm"
            done
        ***REMOVED***
    ***REMOVED***
}

writeReleaseCon***REMOVED***gMapOwnerPatchFile() {
    OWNER_API_VERSION=$1
    OWNER_KIND=$2
    OWNER_NAME=$3
    OWNER_UID=$4
    cat <<EOF > "$OWNER_PATCH_FILE"
{
  "metadata": {
    "ownerReferences": [{
      "apiVersion": "$OWNER_API_VERSION",
      "blockOwnerDeletion": true,
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
    OWNER_API_VERSION=$1
    OWNER_KIND=$2
    OWNER_NAME=$3
    OWNER_UID=$4
    cat <<EOF > "$OWNER_VALUES_FILE"
global:
  ownerReferences:
  - apiVersion: "$OWNER_API_VERSION"
    blockOwnerDeletion: false
    controller: true
    kind: "$OWNER_KIND"
    name: "$OWNER_NAME"
    uid: "$OWNER_UID"
EOF
}

helmUpgrade() {
    RELEASE_NAME=$1
    CHART_LOCATION=$2
    helm upgrade \
        --install \
        --force \
        --namespace "$MY_POD_NAMESPACE" \
        --wait="$HELM_WAIT" \
        --timeout="$HELM_WAIT_TIMEOUT" \
        "$RELEASE_NAME"\
        "$CHART_LOCATION" \
        "${@:3}"
    HELM_EXIT_CODE=$?
    if [ $HELM_EXIT_CODE != 0 ]; then
        echo "helm upgrade failed, exit code: $HELM_EXIT_CODE"
    ***REMOVED***
}

until curl -s $TILLER_READY_ENDPOINT; do
    echo "Waiting for Tiller to become ready"
    sleep 1
done

checkExit

while true; do
    checkExit

    CRD="${HELM_RELEASE_CRD_NAME}.${HELM_RELEASE_CRD_API_GROUP}"
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get "$CRD" \
        -o json > "$HELM_RELEASES_FILE"

    if [ -s "$HELM_RELEASES_FILE" ]; then
        while read -r release; do
            echo -E "$release" > "$CURRENT_RELEASE_FILE"
            RELEASE_NAME="$(jq -Mcr '.metadata.name' "$CURRENT_RELEASE_FILE")"
            RELEASE_UID="$(jq -Mcr '.metadata.uid' "$CURRENT_RELEASE_FILE")"
            RELEASE_API_VERSION="$(jq -Mcr '.apiVersion' "$CURRENT_RELEASE_FILE")"
            RELEASE_RESOURCE_VERSION="$(jq -Mcr '.metadata.resourceVersion' "$CURRENT_RELEASE_FILE")"
            RELEASE_VALUES="$(jq -Mcr '.spec // empty' "$CURRENT_RELEASE_FILE")"
            CHART_LOCATION="$(jq -Mcr '.metadata.annotations["helm-operator.coreos.com/chart-location"] // empty' "$CURRENT_RELEASE_FILE")"

            HELM_ARGS=()
            if [ -s "$EXTRA_VALUES_FILE" ]; then
                HELM_ARGS+=("-f" "$EXTRA_VALUES_FILE")
            ***REMOVED***

            if [ -z "$RELEASE_VALUES" ]; then
                echo "No values, using default values"
            ***REMOVED***
                VALUES_FILE="/tmp/${RELEASE_NAME}-values.yaml"
                echo -E "$RELEASE_VALUES" > "$VALUES_FILE"

                HELM_ARGS+=("-f" "$VALUES_FILE")
            ***REMOVED***

            # If the resource version for this Release CR hasn't changed, we can skip running helm upgrade.
            if [[ -s "/tmp/${RELEASE_NAME}.resourceVersion" && "$(cat "/tmp/${RELEASE_NAME}.resourceVersion")" == "$RELEASE_RESOURCE_VERSION" ]]; then
                echo "Nothing has changed for release $RELEASE_NAME"
            ***REMOVED***
                echo "$RELEASE_RESOURCE_VERSION" > "/tmp/$RELEASE_NAME.resourceVersion"

                writeReleaseOwnerValuesFile "$RELEASE_API_VERSION" "$HELM_RELEASE_CRD_NAME" "$RELEASE_NAME" "$RELEASE_UID"
                writeReleaseCon***REMOVED***gMapOwnerPatchFile "$RELEASE_API_VERSION" "$HELM_RELEASE_CRD_NAME" "$RELEASE_NAME" "$RELEASE_UID"
                HELM_ARGS+=("-f" "$OWNER_VALUES_FILE")

                echo "Running helm upgrade for release $RELEASE_NAME"
                # use the chart location in annotations if speci***REMOVED***ed, otherwise use HELM_CHART_PATH
                CHART="${CHART_LOCATION:-$HELM_CHART_PATH}"
                echo "Using $CHART as chart"
                helmUpgrade "$RELEASE_NAME" "$CHART" "${HELM_ARGS[@]}"

                writeReleaseCon***REMOVED***gmapsFile "$RELEASE_NAME"
                setOwnerOnReleaseCon***REMOVED***gmaps
                cleanupOldReleaseCon***REMOVED***gmaps
            ***REMOVED***

            checkExit
        done < <(jq '.items[]' -Mcr "$HELM_RELEASES_FILE")

        echo "Sleeping $HELM_RECONCILE_INTERVAL_SECONDS seconds"
        for ((i=0; i < $HELM_RECONCILE_INTERVAL_SECONDS; i++)); do
            sleep 1
            checkExit
        done
    ***REMOVED***
        echo "No resources with kind $HELM_RELEASE_CRD_NAME and group $HELM_RELEASE_CRD_API_GROUP"
    ***REMOVED***
done

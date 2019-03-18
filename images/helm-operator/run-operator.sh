#!/bin/bash

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
***REMOVED***

: "${HELM_CHART_PATH:?}"
: "${HELM_RELEASE_CRD_NAME:?}"
: "${HELM_RELEASE_CRD_API_GROUP:?}"

: "${HELM_WAIT:=false}"
: "${HELM_WAIT_TIMEOUT:=120}"
: "${EXTRA_VALUES_FILE:=}"

: "${MY_POD_NAMESPACE:?}"

: "${HELM_RECONCILE_INTERVAL_SECONDS:=120}"
: "${HELM_HOST:="127.0.0.1:44134"}"
: "${TILLER_READY_ENDPOINT:="127.0.0.1:44135/readiness"}"

: "${ALL_NAMESPACES:=false}"
: "${TARGET_NAMESPACES:=$MY_POD_NAMESPACE}"


export HELM_HOST
export RELEASE_HISTORY_LIMIT

NEEDS_EXIT=false

trap setNeedsExit SIGINT SIGTERM

CRD="${HELM_RELEASE_CRD_NAME}.${HELM_RELEASE_CRD_API_GROUP}"
CR_DIRECTORY=/tmp/custom-resources
UPGRADE_RESULT_DIRECTORY=/tmp/helm-upgrade-result
OWNER_PATCH_FILE=/tmp/owner-patch.json
OWNER_VALUES_FILE=/tmp/owner-values.yaml
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
        DELETE_RELEASE_CM_NAMES="$(jq '.items | length as $listLength | ($listLength - (env.RELEASE_HISTORY_LIMIT | tonumber)) as $limitSize | (if $limitSize <= 0 then empty ***REMOVED*** $limitSize end) as $limitSize | sort_by(.metadata.labels.VERSION | tonumber) | limit($limitSize; .[]) | .metadata.name' -rc "$RELEASE_CONFIGMAPS_FILE")"
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
    NAMESPACE=$3
    if helm upgrade \
        --install \
        --namespace "$NAMESPACE" \
        --force \
        --wait="$HELM_WAIT" \
        --timeout="$HELM_WAIT_TIMEOUT" \
        "$RELEASE_NAME"\
        "$CHART_LOCATION" \
        "${@:4}";
    then
        echo "Helm upgrade succeeded"
        return 0
    ***REMOVED***
        echo "Helm upgrade failed"
        return 1
    ***REMOVED***
}

until curl -s $TILLER_READY_ENDPOINT; do
    echo "Waiting for Tiller to become ready"
    sleep 1
done

checkExit

echo "Target namespaces: $TARGET_NAMESPACES"

TARGET_NAMESPACES_LIST=()
while read -rd, ns; do
    TARGET_NAMESPACES_LIST+=("$ns")
done <<<"$TARGET_NAMESPACES,"

while true; do
    checkExit

    rm -rf "$CR_DIRECTORY"
    mkdir -p "$CR_DIRECTORY"

    CR_NAMES_LIST=()

    # Gather all the CR instances for the CRD we're watching
    if [ "$ALL_NAMESPACES" == "true" ]; then
        echo "Querying all namespaces for $CRD resources"
        CURRENT_CR_LIST="/tmp/cr-list.json"
        rm -f "$CURRENT_CR_LIST"
        kubectl --all-namespaces get "$CRD" -o json > "$CURRENT_CR_LIST"
        echo "Got $(jq -Mcr '.items | length' "$CURRENT_CR_LIST") instances of $CRD from all namespaces"
        while read -r cr_content; do
            NAME="$(echo "$cr_content" | jq -Mcr '.metadata.name')"
            NAMESPACE="$(echo "$cr_content" | jq -Mcr '.metadata.namespace')"
            PROCESS_CR=false

            if [ ${#TARGET_NAMESPACES_LIST[@]} -eq 0 ]; then
                PROCESS_CR=true
            ***REMOVED***

            for TARGET_NAMESPACE in "${TARGET_NAMESPACES_LIST[@]}"; do
                if [ "$NAMESPACE" == "$TARGET_NAMESPACE" ]; then
                    PROCESS_CR=true
                    break
                ***REMOVED***
            done

            if [ "$PROCESS_CR" == "true" ]; then
                CR_NAMES_LIST+=("$NAMESPACE/$NAME")
                CR_FILE="${CR_DIRECTORY}/${NAMESPACE}-${NAME}-cr.json"
                echo "$cr_content" > "${CR_FILE}"
            ***REMOVED***

        done < <(jq -Mcr '.items[]' "$CURRENT_CR_LIST")
    ***REMOVED***
        for TARGET_NAMESPACE in "${TARGET_NAMESPACES_LIST[@]}"; do
            echo "Querying $TARGET_NAMESPACE for $CRD resources"
            CURRENT_CR_LIST="/tmp/cr-list.json"
            rm -f "$CURRENT_CR_LIST"
            kubectl --namespace "$TARGET_NAMESPACE" get "$CRD" -o json > "$CURRENT_CR_LIST"
            echo "Got $(jq -r '.items | length' "$CURRENT_CR_LIST") instances of $CRD from $TARGET_NAMESPACE"
            while read -r cr_content; do
                NAME="$(echo "$cr_content" | jq -Mcr '.metadata.name')"
                NAMESPACE="$(echo "$cr_content" | jq -Mcr '.metadata.namespace')"
                CR_NAMES_LIST+=("$NAMESPACE/$NAME")
                CR_FILE="${CR_DIRECTORY}/${NAMESPACE}-${NAME}-cr.json"
                echo "$cr_content" > "${CR_FILE}"
            done < <(jq -Mcr '.items[]' "$CURRENT_CR_LIST")
        done
    ***REMOVED***

    checkExit

    echo "Got ${#CR_NAMES_LIST[@]} total instances of ${CRD}: ${CR_NAMES_LIST[*]}"

    ***REMOVED***nd "$CR_DIRECTORY" -type f -name '*.json' | while read -r CR_FILE; do
        RESOURCE_KIND="$HELM_RELEASE_CRD_NAME"
        RESOURCE_NAME="$(jq -Mcr '.metadata.name' "$CR_FILE")"
        RESOURCE_NAMESPACE="$(jq -Mcr '.metadata.namespace' "$CR_FILE")"
        FULL_RESOURCE_NAME="$RESOURCE_NAMESPACE/$RESOURCE_NAME"
        RESOURCE_UID="$(jq -Mcr '.metadata.uid' "$CR_FILE")"
        RESOURCE_API_VERSION="$(jq -Mcr '.apiVersion' "$CR_FILE")"
        RESOURCE_RESOURCE_VERSION="$(jq -Mcr '.metadata.resourceVersion' "$CR_FILE")"
        RESOURCE_VALUES="$(jq -Mcr '.spec // {}' "$CR_FILE")"
        CHART_LOCATION="$(jq -Mcr '.metadata.annotations["helm-operator.coreos.com/chart-location"] // empty' "$CR_FILE")"
        # use the chart location in annotations if speci***REMOVED***ed, otherwise use HELM_CHART_PATH
        CHART="${CHART_LOCATION:-$HELM_CHART_PATH}"

        RELEASE_NAME=""
        # If running against all namespaces, we have to pre***REMOVED***x the namespace to
        # the release name because helm doesn't namespace releases by
        # namespace.
        if [ "$ALL_NAMESPACES" == "true" ]; then
            RELEASE_NAME="$RESOURCE_NAMESPACE-$RESOURCE_NAME"
        ***REMOVED***
            RELEASE_NAME="$RESOURCE_NAME"
        ***REMOVED***

        if [ ${#RELEASE_NAME} -gt 52 ]; then
            echo "$CRD generated release name cannot be more than than 52 characters, got ${#RELEASE_NAME} for $RELEASE_NAME"
            continue
        ***REMOVED***
        echo "Processing $CRD $FULL_RESOURCE_NAME at resourceVersion: $RESOURCE_RESOURCE_VERSION using $CHART as chart and $RELEASE_NAME as release name"

        HELM_ARGS=()
        if [ -s "$EXTRA_VALUES_FILE" ]; then
            HELM_ARGS+=("-f" "$EXTRA_VALUES_FILE")
        ***REMOVED***

        VALUES_FILE="/tmp/$RESOURCE_NAMESPACE-$RESOURCE_NAME-values.yaml"
        NEW_VALUES_FILE="/tmp/$RESOURCE_NAMESPACE-$RESOURCE_NAME-values-NEW.yaml"

        echo -E "$RESOURCE_VALUES" > "$NEW_VALUES_FILE"

        # If the spec for this CR hasn't changed, we can skip running helm upgrade.
        if [[ -a "$VALUES_FILE" && "$(cat "$VALUES_FILE")" == "$(cat "$NEW_VALUES_FILE")" ]]; then
            echo "Nothing has changed for $FULL_RESOURCE_NAME"
        ***REMOVED***
            if [ -a "$VALUES_FILE" ]; then
                echo "$CRD $FULL_RESOURCE_NAME has been modi***REMOVED***ed"
            ***REMOVED***
                echo "New $CRD $FULL_RESOURCE_NAME"
            ***REMOVED***
            HELM_ARGS+=("-f" "$NEW_VALUES_FILE")

            writeReleaseOwnerValuesFile "$RESOURCE_API_VERSION" "$RESOURCE_KIND" "$RESOURCE_NAME" "$RESOURCE_UID"
            writeReleaseCon***REMOVED***gMapOwnerPatchFile "$RESOURCE_API_VERSION" "$RESOURCE_KIND" "$RESOURCE_NAME" "$RESOURCE_UID"
            HELM_ARGS+=("-f" "$OWNER_VALUES_FILE")

            mkdir -p "$UPGRADE_RESULT_DIRECTORY"
            echo "Running helm upgrade for release $RELEASE_NAME"
            if helmUpgrade "$RELEASE_NAME" "$CHART" "$RESOURCE_NAMESPACE" "${HELM_ARGS[@]}"; then
                # store the last version that we were able to process so we can
                # compare new specs against it later
                mv -f "$NEW_VALUES_FILE" "$VALUES_FILE"
                echo "Updating $CRD $FULL_RESOURCE_NAME status"
                kubectl \
                    -n "$RESOURCE_NAMESPACE" \
                    patch "$RESOURCE_KIND" "$RESOURCE_NAME" \
                    --type json \
                    -p '[{"op": "add", "path": "/status", "value":{}},{"op": "add", "path": "/status/observedVersion", "value":"'"$RESOURCE_RESOURCE_VERSION"'"}]'

                writeReleaseCon***REMOVED***gmapsFile "$RELEASE_NAME"
                setOwnerOnReleaseCon***REMOVED***gmaps
                cleanupOldReleaseCon***REMOVED***gmaps
            ***REMOVED***
                echo "Error occurred when processing $FULL_RESOURCE_NAME"
            ***REMOVED***
        ***REMOVED***
        checkExit
    done

    checkExit

    echo "Sleeping $HELM_RECONCILE_INTERVAL_SECONDS seconds"
    for ((i=0; i < $HELM_RECONCILE_INTERVAL_SECONDS; i++)); do
        sleep 1
        checkExit
    done
done

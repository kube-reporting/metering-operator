#!/bin/bash

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
fi

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
RESOURCE_VERSIONS_DIRECTORY=/tmp/resource-versions
OWNER_PATCH_FILE=/tmp/owner-patch.json
OWNER_VALUES_FILE=/tmp/owner-values.yaml
RELEASE_CONFIGMAPS_FILE=/tmp/release-configmaps.json

setNeedsExit() {
    echo "Got shutdown signal"
    NEEDS_EXIT=true
}

checkExit() {
    if [ "$NEEDS_EXIT" == "true" ]; then
        echo "finished shutdown, exiting"
        exit 0
    fi
}

writeReleaseConfigmapsFile() {
    HELM_RELEASE_NAME=$1
    kubectl \
        --namespace "$MY_POD_NAMESPACE" \
        get configmap \
        -l "OWNER=TILLER,NAME=$HELM_RELEASE_NAME" \
        -o json | jq '.' -r > "$RELEASE_CONFIGMAPS_FILE"
}

setOwnerOnReleaseConfigmaps(){
    if [ "$SET_OWNER_REFERENCE_VALUE" == "true" ]; then
        echo "Setting ownerReferences for Helm release configmaps"

        RELEASE_CM_NAMES="$(jq '.items[] | select(.metadata.ownerReferences | length == 0) | .metadata.name' -r "$RELEASE_CONFIGMAPS_FILE")"
        if [ -z "$RELEASE_CM_NAMES" ]; then
            echo "No release configmaps to patch ownership of yet"
        else
            echo "$RELEASE_CM_NAMES" | while read -r cm; do
                echo "Setting owner of $cm"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    patch configmap "$cm" \
                    -p "$(cat $OWNER_PATCH_FILE)"
            done
        fi
    fi
}

cleanupOldReleaseConfigmaps() {
    if [ -n "$RELEASE_HISTORY_LIMIT" ]; then
        echo "Getting list of helm release configmaps to delete"
        DELETE_RELEASE_CM_NAMES="$(jq '.items | length as $listLength | ($listLength - (env.RELEASE_HISTORY_LIMIT | tonumber)) as $limitSize | (if $limitSize <= 0 then empty else $limitSize end) as $limitSize | sort_by(.metadata.labels.VERSION | tonumber) | limit($limitSize; .[]) | .metadata.name' -rc "$RELEASE_CONFIGMAPS_FILE")"
        if [ -z "$DELETE_RELEASE_CM_NAMES" ]; then
            echo "No release configmaps to delete yet"
        else
            echo "$DELETE_RELEASE_CM_NAMES" | while read -r cm; do
                echo "Deleting helm release configmap $cm"
                kubectl \
                    --namespace "$MY_POD_NAMESPACE" \
                    delete configmap "$cm"
            done
        fi
    fi
}

writeReleaseConfigMapOwnerPatchFile() {
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
    helm upgrade \
        --install \
        --namespace "$NAMESPACE" \
        --wait="$HELM_WAIT" \
        --force \
        --timeout="$HELM_WAIT_TIMEOUT" \
        "$RELEASE_NAME"\
        "$CHART_LOCATION" \
        "${@:4}"
    HELM_EXIT_CODE=$?
    if [ $HELM_EXIT_CODE != 0 ]; then
        echo "helm upgrade failed, exit code: $HELM_EXIT_CODE"
    fi
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
            fi

            for TARGET_NAMESPACE in "${TARGET_NAMESPACES_LIST[@]}"; do
                if [ "$NAMESPACE" == "$TARGET_NAMESPACE" ]; then
                    PROCESS_CR=true
                    break
                fi
            done

            if [ "$PROCESS_CR" == "true" ]; then
                CR_NAMES_LIST+=("$NAMESPACE/$NAME")
                CR_FILE="${CR_DIRECTORY}/${NAMESPACE}-${NAME}-cr.json"
                echo "$cr_content" > "${CR_FILE}"
            fi

        done < <(jq -Mcr '.items[]' "$CURRENT_CR_LIST")
    else
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
    fi

    checkExit

    echo "Got ${#CR_NAMES_LIST[@]} total instances of ${CRD}: ${CR_NAMES_LIST[*]}"

    find "$CR_DIRECTORY" -type f -name '*.json' | while read -r CR_FILE; do
        RESOURCE_KIND="$HELM_RELEASE_CRD_NAME"
        RESOURCE_NAME="$(jq -Mcr '.metadata.name' "$CR_FILE")"
        RESOURCE_NAMESPACE="$(jq -Mcr '.metadata.namespace' "$CR_FILE")"
        FULL_RESOURCE_NAME="$RESOURCE_NAMESPACE/$RESOURCE_NAME"
        RESOURCE_UID="$(jq -Mcr '.metadata.uid' "$CR_FILE")"
        RESOURCE_API_VERSION="$(jq -Mcr '.apiVersion' "$CR_FILE")"
        RESOURCE_RESOURCE_VERSION="$(jq -Mcr '.metadata.resourceVersion' "$CR_FILE")"
        RESOURCE_VALUES="$(jq -Mcr '.spec // empty' "$CR_FILE")"
        CHART_LOCATION="$(jq -Mcr '.metadata.annotations["helm-operator.coreos.com/chart-location"] // empty' "$CR_FILE")"
        # use the chart location in annotations if specified, otherwise use HELM_CHART_PATH
        CHART="${CHART_LOCATION:-$HELM_CHART_PATH}"

        RELEASE_NAME=""
        # If running against all namespaces, we have to prefix the namespace to
        # the release name because helm doesn't namespace releases by
        # namespace.
        if [ "$ALL_NAMESPACES" == "true" ]; then
            RELEASE_NAME="$RESOURCE_NAMESPACE-$RESOURCE_NAME"
        else
            RELEASE_NAME="$RESOURCE_NAME"
        fi

        if [ ${#RELEASE_NAME} -gt 52 ]; then
            echo "$CRD generated release name cannot be more than than 52 characters, got ${#RELEASE_NAME} for $RELEASE_NAME"
            continue
        fi
        echo "Processing $CRD $FULL_RESOURCE_NAME at resourceVersion: $RESOURCE_RESOURCE_VERSION using $CHART as chart and $RELEASE_NAME as release name"

        HELM_ARGS=()
        if [ -s "$EXTRA_VALUES_FILE" ]; then
            HELM_ARGS+=("-f" "$EXTRA_VALUES_FILE")
        fi

        if [ -z "$RESOURCE_VALUES" ]; then
            echo "No values for $FULL_RESOURCE_NAME, using default values"
        else
            VALUES_FILE="/tmp/$RESOURCE_NAMESPACE-$RESOURCE_NAME-values.yaml"
            echo -E "$RESOURCE_VALUES" > "$VALUES_FILE"

            HELM_ARGS+=("-f" "$VALUES_FILE")
        fi

        mkdir -p "$RESOURCE_VERSIONS_DIRECTORY"
        RESOURCE_VERSION_FILE="$RESOURCE_VERSIONS_DIRECTORY/$RESOURCE_NAMESPACE-$RESOURCE_NAME.resourceVersion"
        # If the resource version for this Release CR hasn't changed, we can skip running helm upgrade.
        if [[ -s "$RESOURCE_VERSION_FILE" && "$(cat "$RESOURCE_VERSION_FILE")" == "$RESOURCE_RESOURCE_VERSION" ]]; then
            echo "Nothing has changed for $FULL_RESOURCE_NAME"
        else
            echo "$CRD $FULL_RESOURCE_NAME has been modified"
            echo "$RESOURCE_RESOURCE_VERSION" > "$RESOURCE_VERSION_FILE"

            writeReleaseOwnerValuesFile "$RESOURCE_API_VERSION" "$RESOURCE_KIND" "$RESOURCE_NAME" "$RESOURCE_UID"
            writeReleaseConfigMapOwnerPatchFile "$RESOURCE_API_VERSION" "$RESOURCE_KIND" "$RESOURCE_NAME" "$RESOURCE_UID"
            HELM_ARGS+=("-f" "$OWNER_VALUES_FILE")

            mkdir -p "$UPGRADE_RESULT_DIRECTORY"
            echo "Running helm upgrade for release $RELEASE_NAME"
            if helmUpgrade "$RELEASE_NAME" "$CHART" "$RESOURCE_NAMESPACE" "${HELM_ARGS[@]}" | tee "$UPGRADE_RESULT_DIRECTORY/$RELEASE_NAME.txt"; then
                echo "Error occurred when processing $FULL_RESOURCE_NAME"
            fi

            writeReleaseConfigmapsFile "$RESOURCE_NAME"
            setOwnerOnReleaseConfigmaps
            cleanupOldReleaseConfigmaps
        fi
        checkExit
    done

    checkExit

    echo "Sleeping $HELM_RECONCILE_INTERVAL_SECONDS seconds"
    for ((i=0; i < $HELM_RECONCILE_INTERVAL_SECONDS; i++)); do
        sleep 1
        checkExit
    done
done

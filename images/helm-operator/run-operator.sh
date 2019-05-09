#!/bin/bash

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
fi

: "${HELM_CHART_PATH:?}"
: "${HELM_RELEASE_CRD_NAME:?}"
: "${HELM_RELEASE_CRD_API_GROUP:?}"

: "${EXTRA_VALUES_FILE:=}"

: "${MY_POD_NAMESPACE:?}"

: "${HELM_RECONCILE_INTERVAL_SECONDS:=120}"

: "${ALL_NAMESPACES:=false}"
: "${TARGET_NAMESPACES:=$MY_POD_NAMESPACE}"


export HELM_HOST

NEEDS_EXIT=false

trap setNeedsExit SIGINT SIGTERM

CRD="${HELM_RELEASE_CRD_NAME}.${HELM_RELEASE_CRD_API_GROUP}"
CR_DIRECTORY=/tmp/custom-resources
UPGRADE_RESULT_DIRECTORY=/tmp/helm-upgrade-result

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

deployResources() {
    RELEASE_NAME=$1
    CHART_LOCATION=$2
    NAMESPACE=$3
    OWNER_API_VERSION=${4:?}
    OWNER_KIND=${5:?}
    OWNER_NAME=${6:?}
    OWNER_UID=${7:?}
    FLAGS=("${@:8}")

    PRUNE_LABEL_KEY='metering.openshift.io/prune'

    if deploy-resources.sh \
        "${FLAGS[@]}" \
        "$RELEASE_NAME"\
        "$CHART_LOCATION" \
        "$NAMESPACE" \
        "$OWNER_API_VERSION" \
        "$OWNER_KIND" \
        "$OWNER_NAME" \
        "$OWNER_UID" \
        "$PRUNE_LABEL_KEY" \
        false;
    then
        echo "Deploy resources succeeded"
        return 0
    else
        echo "Deploy resources failed"
        return 1
    fi
}

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
        RESOURCE_OBSERVED_VERSION="$(jq -Mcr '.status.observedVersion' "$CR_FILE")"
        RESOURCE_VALUES="$(jq -Mcr '.spec // {}' "$CR_FILE")"
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

        DEPLOY_RESOURCES_FLAGS=()
        if [ -s "$EXTRA_VALUES_FILE" ]; then
            DEPLOY_RESOURCES_FLAGS+=("-f" "$EXTRA_VALUES_FILE")
        fi

        VALUES_FILE="/tmp/$RESOURCE_NAMESPACE-$RESOURCE_NAME-values.yaml"
        NEW_VALUES_FILE="/tmp/$RESOURCE_NAMESPACE-$RESOURCE_NAME-values-NEW.yaml"

        echo -E "$RESOURCE_VALUES" > "$NEW_VALUES_FILE"

        # If the spec for this CR hasn't changed, we can skip redeploying
        if [[ -a "$VALUES_FILE" && "$(cat "$VALUES_FILE")" == "$(cat "$NEW_VALUES_FILE")" ]]; then
            echo "Nothing has changed for $FULL_RESOURCE_NAME"
        else
            if [ -a "$VALUES_FILE" ]; then
                echo "$CRD $FULL_RESOURCE_NAME has been modified"
            else
                echo "New $CRD $FULL_RESOURCE_NAME"
            fi
            DEPLOY_RESOURCES_FLAGS+=("-f" "$NEW_VALUES_FILE")

            if [ "$RESOURCE_OBSERVED_VERSION" != "null" ]; then
                DEPLOY_RESOURCES_FLAGS+=(--is-upgrade)
            fi

            mkdir -p "$UPGRADE_RESULT_DIRECTORY"
            echo "Deploying resources for release $RELEASE_NAME"
            if deployResources \
                "$RELEASE_NAME" "$CHART" "$RESOURCE_NAMESPACE" \
                "$RESOURCE_API_VERSION" "$RESOURCE_KIND" "$RESOURCE_NAME" "$RESOURCE_UID" "${DEPLOY_RESOURCES_FLAGS[@]}"; then
                # store the last version that we were able to process so we can
                # compare new specs against it later
                mv -f "$NEW_VALUES_FILE" "$VALUES_FILE"
                echo "Updating $CRD $FULL_RESOURCE_NAME status"
                kubectl \
                    -n "$RESOURCE_NAMESPACE" \
                    patch "$RESOURCE_KIND" "$RESOURCE_NAME" \
                    --type json \
                    -p '[{"op": "add", "path": "/status", "value":{}},{"op": "add", "path": "/status/observedVersion", "value":"'"$RESOURCE_RESOURCE_VERSION"'"}]'
            else
                echo "Error occurred when processing $FULL_RESOURCE_NAME"
            fi
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

#!/bin/bash

CUR_DIR=$(dirname "${BASH_SOURCE[0]}")

set -eE -o functrace
set -u

shopt -s extglob


: "${ENABLE_DEBUG:=false}"

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
fi

: "${KUBECTL_BIN:=kubectl}"
: "${OC_BIN:=oc}"
: "${HELM_TEMPLATE_CMD:="$CUR_DIR/helm-template.sh"}"
: "${KUBECTL_APPLY_PRUNE_CMD:="$CUR_DIR/kubectl-apply-prune.sh"}"
: "${ADD_LABEL_BIN:="$CUR_DIR/add-label.sh"}"
: "${ADD_OWNER_REF_BIN:="$CUR_DIR/add-owner-ref.sh"}"
: "${FAQ_BIN:=faq}"

: "${ENABLE_OWNER_REFERENCES=true}"

export KUBECTL_BIN
export OC_BIN
export FAQ_BIN

VALUES_FILES=()
IS_UPGRADE=false

die() {
    printf '%s\n' "$1" >&2
    exit 1
}

while :; do
    case $1 in
        -f|--values)       # Takes an option argument; ensure it has been specified.
            if [ "$2" ]; then
                VALUES_FILES+=("$2")
                shift
            else
                die 'ERROR: "-f" requires a non-empty option argument.'
            fi
            ;;
        --is-upgrade)
            IS_UPGRADE=true
            ;;
        --)              # End of all options.
            shift
            break
            ;;
        -?*)
            printf 'WARN: Unknown option (ignored): %s\n' "$1" >&2
            ;;
        *)               # Default case: No more options, so break out of the loop.
            break
    esac

    shift
done

RELEASE_NAME=${1:?}
CHART=${2:?}
NAMESPACE=${3:?}

if [ "$ENABLE_OWNER_REFERENCES" == "true" ]; then
    OWNER_API_VERSION=${4:?}
    OWNER_KIND=${5:?}
    OWNER_NAME=${6:?}
    OWNER_UID=${7:?}
    shift 4
else
    OWNER_API_VERSION=""
    OWNER_KIND=""
    OWNER_NAME=""
    OWNER_UID=""
fi

PRUNE_LABEL_KEY=${4:?}
DRY_RUN=${5:?}

TMP_DIR="$(mktemp -d)"

cleanup() {
    # preserve original exit code
    exit_status=$?
    # kill background jobs
    JOBS="$(jobs -p)"
    if [ -n "${JOBS}" ]; then
        echo "Stopping background jobs"
        kill -KILL "${JOBS}" 2> /dev/null || true > /dev/null
        echo "Waiting for background jobs"
        # Wait for any jobs
        wait 2>/dev/null
    fi
    rm -rf "$TMP_DIR"
    # exit
    exit "$exit_status"
}
trap cleanup EXIT

# Get the chart defaults
helm inspect values "$CHART" > "$TMP_DIR/default-values.yaml"

# merge the provided values with the defaults by slurping each values file into
# a single array, and using reduce to apply the `*` to do a recursive merge of
# each object in the array

# shellcheck disable=SC2016
"$FAQ_BIN" --slurp -f yaml -o yaml -r -M -c 'reduce .[] as $item ({}; . * $item)' "$TMP_DIR/default-values.yaml" ${VALUES_FILES[@]+"${VALUES_FILES[@]}"} > "$TMP_DIR/merged-values.yaml"

# gets the value from the merged values file that will be used for helm templating
getHelmValue() {
    JQ_PROG=$1
    "$FAQ_BIN" -f yaml -o yaml -r -M -c "$JQ_PROG" "$TMP_DIR/merged-values.yaml"
}

# the variables below are used to determine if we need to create/update or delete a resource that isn't always created.

# openshift-metering chart values
: "${CREATE_METERING_DEFAULT_STORAGE:="$(getHelmValue '.defaultStorage.create')"}"
: "${CREATE_METERING_MONITORING_RESOURCES:="$(getHelmValue '.monitoring.enabled')"}"
: "${CREATE_METERING_MONITORING_RBAC:="$(getHelmValue '.monitoring.enabled and .monitoring.enabled and .monitoring.createRBAC')"}"
: "${ENABLE_REPORTING_AWS_BILLING:="$(getHelmValue '.awsBillingReportDataSource.enabled')"}"

# hdfs chart values
: "${ENABLE_HDFS:="$(getHelmValue '.hdfs.enabled')"}"

# presto chart values
: "${CREATE_HIVE_METASTORE_PVC:="$(getHelmValue '.presto.spec.hive.metastore.storage.create')"}"
: "${CREATE_PRESTO_SHARED_VOLUME_PVC:="$(getHelmValue '.presto.spec.config.sharedVolume.enabled and .presto.spec.config.sharedVolume.createPVC')"}"
: "${CREATE_PRESTO_AWS_CREDENTIALS:="$(getHelmValue '.presto.spec.config.createAwsCredentialsSecret')"}"

# reporting-operator chart values
: "${CREATE_REPORTING_OPERATOR_AUTH_PROXY_COOKIE_SECRET:="$(getHelmValue '.["reporting-operator"].spec.authProxy.enabled and .["reporting-operator"].spec.authProxy.createCookieSecret')"}"
: "${CREATE_REPORTING_OPERATOR_AUTH_PROXY_HTPASSWD_SECRET:="$(getHelmValue '.["reporting-operator"].spec.authProxy.enabled and .["reporting-operator"].spec.authProxy.createHtpasswdSecret')"}"
: "${CREATE_REPORTING_OPERATOR_AUTH_PROXY_AUTHENTICATED_EMAILS_SECRET:="$(getHelmValue '.["reporting-operator"].spec.authProxy.enabled and .["reporting-operator"].spec.authProxy.createAuthenticatedEmailsSecret')"}"
: "${CREATE_REPORTING_OPERATOR_AUTH_PROXY_RBAC:="$(getHelmValue '.["reporting-operator"].spec.authProxy.enabled and (.["reporting-operator"].spec.authProxy.subjectAccessReviewEnabled and .["reporting-operator"].spec.authProxy.delegateURLsEnabled) and .["reporting-operator"].spec.authProxy.createAuthProxyClusterRole')"}"
: "${CREATE_REPORTING_OPERATOR_PROMETHEUS_BEARER_TOKEN:="$(getHelmValue '.["reporting-operator"].spec.config.prometheusImporter.auth.tokenSecret.create')"}"
: "${CREATE_REPORTING_OPERATOR_PROMETHEUS_CERTIFICATE_AUTHORITY:="$(getHelmValue '.["reporting-operator"].spec.config.prometheusCertificateAuthority.configMap.create')"}"
: "${CREATE_REPORTING_OPERATOR_AWS_CREDENTIALS:="$(getHelmValue '.["reporting-operator"].spec.config.createAwsCredentialsSecret')"}"
: "${CREATE_REPORTING_OPERATOR_TLS_SECRETS:="$(getHelmValue '.["reporting-operator"].spec.config.tls.createSecret or .["reporting-operator"].spec.config.metricsTLS.createSecret')"}"
: "${CREATE_REPORTING_OPERATOR_ROUTE:="$(getHelmValue '.["reporting-operator"].spec.route.enabled')"}"
: "${CREATE_REPORTING_OPERATOR_CLUSTER_MONITORING_VIEW_RBAC:="$(getHelmValue '.["reporting-operator"].spec.config.createClusterMonitoringViewRBAC')"}"


addLabel() {
    local PRUNE_LABEL_VALUE=$1
    "$ADD_LABEL_BIN" "$PRUNE_LABEL_KEY" "$PRUNE_LABEL_VALUE"
}

addOwnerRef() {
    local BLOCK_OWNER_DELETION="$1"
    "$ADD_OWNER_REF_BIN" "$OWNER_API_VERSION" "$OWNER_KIND" "$OWNER_NAME" "$OWNER_UID" "$BLOCK_OWNER_DELETION"
}

kubectlApply() {
    local KUBECTL_ARGS=(\
        --namespace "$NAMESPACE" \
    )
    if [ "$DRY_RUN" == "true" ]; then
        KUBECTL_ARGS+=(--dry-run)
    fi
    if [ "$ENABLE_DEBUG" == "true" ]; then
        KUBECTL_ARGS+=(-o yaml)
    fi

    "$KUBECTL_BIN" apply "${KUBECTL_ARGS[@]}" -f -
}

helmTemplate() {
    local FILE=$1
    local HELM_TEMPLATE_ARGS=( "$RELEASE_NAME" "$CHART" "$NAMESPACE"  -x "$FILE" )

    for valuesFile in ${VALUES_FILES[@]+"${VALUES_FILES[@]}"}; do
        HELM_TEMPLATE_ARGS+=(-f "$valuesFile")
    done

    "$HELM_TEMPLATE_CMD" "${HELM_TEMPLATE_ARGS[@]}"
}

helmTemplateAndApply() {
    local FILE="$1"
    local PRUNE_LABEL_VALUE="$2"
    local SET_OWNER="$3"
    local BLOCK_OWNER_DELETION="$4"

    local OUTPUT
    OUTPUT="$(helmTemplate "$FILE" | addLabel "$PRUNE_LABEL_VALUE")"

    if [ "$ENABLE_OWNER_REFERENCES" == "true" ] && [ "$SET_OWNER" == "true" ]; then
        OUTPUT="$(echo "$OUTPUT" | addOwnerRef "$BLOCK_OWNER_DELETION")"
    fi

    echo "$OUTPUT" | kubectlApply
}

kubectlDeleteByPruneLabelValue() {
    local KINDS=$1
    local PRUNE_LABEL_VALUE=$2
    local SELECTOR="$PRUNE_LABEL_KEY=$PRUNE_LABEL_VALUE"
    local NAMES=()
    while IFS='' read -r line; do
        NAMES+=("$line");
    done < <("$KUBECTL_BIN" --namespace "$NAMESPACE" get "${KINDS}" -l "$SELECTOR" -o name)

    if [[ ${#NAMES[@]} -eq 0 ]]; then
        echo "No $KINDS resources to delete based on selector $SELECTOR"
        return
    fi
    if [ "$DRY_RUN" == "true" ]; then
        for resource in "${NAMES[@]}"; do
            echo "$resource deleted (dry run)"
        done
    else
        echo "Deleting ${NAMES[*]}"
        "$KUBECTL_BIN" \
            --namespace "$NAMESPACE"\
            delete \
            "${KINDS}" \
            -l "$SELECTOR"
    fi
}

installMeteringResources() {
    if [ "$CREATE_METERING_DEFAULT_STORAGE" == "true" ]; then
        helmTemplateAndApply templates/metering/default-storage-location.yaml openshift-metering-default-storage-location true false
    else
        kubectlDeleteByPruneLabelValue storagelocation openshift-metering-default-storage-location
    fi

    if [ "$CREATE_METERING_MONITORING_RBAC" == "true" ]; then
        helmTemplateAndApply templates/monitoring/monitoring-rbac.yaml openshift-metering-monitoring-rbac true false
    else
        kubectlDeleteByPruneLabelValue role,rolebinding openshift-metering-monitoring-rbac
    fi

    if [ "$CREATE_METERING_MONITORING_RESOURCES" == "true" ]; then
        helmTemplateAndApply templates/monitoring/presto-service-monitor.yaml openshift-metering-presto-service-monitor true false
        helmTemplateAndApply templates/monitoring/reporting-operator-service-monitor.yaml openshift-metering-reporting-operator-service-monitor true false
    else
        kubectlDeleteByPruneLabelValue servicemonitor openshift-metering-presto-service-monitor
        kubectlDeleteByPruneLabelValue servicemonitor openshift-metering-reporting-operator-service-monitor
        kubectlDeleteByPruneLabelValue role,rolebinding openshift-metering-monitoring-rbac
    fi

    helmTemplateAndApply templates/metering/metering-roles.yaml openshift-metering-roles true false
    helmTemplateAndApply templates/metering/metering-rolebindings.yaml openshift-metering-rolebindings true false

}

installReportingResources() {
    helmTemplateAndApply templates/openshift-reporting/datasources/default-datasources.yaml default-datasources true false
    if [ "$ENABLE_REPORTING_AWS_BILLING" == "true" ]; then
        helmTemplateAndApply templates/openshift-reporting/datasources/aws-datasources.yaml aws-datasources true false
        helmTemplateAndApply templates/openshift-reporting/report-queries/aws-billing.yaml report-queries-aws-billing true false
        helmTemplateAndApply templates/openshift-reporting/report-queries/pod-cpu-aws.yaml report-queries-pod-cpu-aws true false
        helmTemplateAndApply templates/openshift-reporting/report-queries/pod-memory-aws.yaml report-queries-pod-memory-aws true false
    else
        kubectlDeleteByPruneLabelValue datasources aws-datasources
        kubectlDeleteByPruneLabelValue reportquery report-queries-aws-billing
        kubectlDeleteByPruneLabelValue reportquery report-queries-pod-cpu-aws
        kubectlDeleteByPruneLabelValue reportquery report-queries-pod-memory-aws
    fi

    helmTemplateAndApply templates/openshift-reporting/report-queries/cluster-capacity.yaml report-queries-cluster-capacity true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/cluster-usage.yaml report-queries-cluster-usage true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/cluster-utilization.yaml report-queries-cluster-utilization true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/node-cpu.yaml report-queries-node-cpu true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/node-memory.yaml report-queries-node-memory true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/persistentvolumeclaim-capacity.yaml report-queries-persistentvolumeclaim-capacity true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/persistentvolumeclaim-request.yaml report-queries-persistentvolumeclaim-request true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/persistentvolumeclaim-usage.yaml report-queries-persistentvolumeclaim-usage true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/pod-cpu.yaml report-queries-pod-cpu true false
    helmTemplateAndApply templates/openshift-reporting/report-queries/pod-memory.yaml report-queries-pod-memory true false
}

installHdfsResources() {
    if [ "$ENABLE_HDFS" == "true" ]; then
        helmTemplateAndApply templates/hdfs/hdfs-configmap.yaml hdfs-configmap true false
        helmTemplateAndApply templates/hdfs/hdfs-serviceaccount.yaml hdfs-service-account true false
        helmTemplateAndApply templates/hdfs/hdfs-datanode-statefulset.yaml hdfs-datanode-statefulset true false
        helmTemplateAndApply templates/hdfs/hdfs-namenode-statefulset.yaml hdfs-namenode-statefulset true false
    else
        kubectlDeleteByPruneLabelValue configmap hdfs-configmap
        kubectlDeleteByPruneLabelValue serviceaccount hdfs-service-account
        kubectlDeleteByPruneLabelValue statefulset hdfs-datanode-statefulset
        kubectlDeleteByPruneLabelValue statefulset hdfs-namenode-statefulset
    fi
}

installPrestoResources() {
    if [ "$CREATE_HIVE_METASTORE_PVC" == "true" ]; then
        if [ "$IS_UPGRADE" == "false" ]; then
            helmTemplateAndApply templates/presto/hive-metastore-pvc.yaml hive-metastore-pvc true false
        fi
    else
        kubectlDeleteByPruneLabelValue persistentvolumeclaim hive-metastore-pvc
    fi

    if [ "$CREATE_PRESTO_SHARED_VOLUME_PVC" == "true" ]; then
        helmTemplateAndApply templates/presto/shared-volume-pvc.yaml presto-shared-volume-pvc true false
    else
        kubectlDeleteByPruneLabelValue persistentvolumeclaim presto-shared-volume-pvc
    fi

    helmTemplateAndApply templates/presto/hive-configmap.yaml hive-configmap true false
    helmTemplateAndApply templates/presto/hive-scripts-configmap.yaml hive-scripts-configmap true false
    helmTemplateAndApply templates/presto/hive-metastore-service.yaml hive-metastore-service true false
    helmTemplateAndApply templates/presto/hive-metastore-statefulset.yaml hive-metastore-statefulset true false
    helmTemplateAndApply templates/presto/hive-server-service.yaml hive-server-service true false
    helmTemplateAndApply templates/presto/hive-serviceaccount.yaml hive-serviceaccount true false
    helmTemplateAndApply templates/presto/hive-server-statefulset.yaml hive-server-statefulset true false

    if [ "$CREATE_PRESTO_AWS_CREDENTIALS" == "true" ]; then
        helmTemplateAndApply templates/presto/presto-aws-credentials-secret.yaml presto-aws-credentials-secret true false
    else
        kubectlDeleteByPruneLabelValue secret presto-aws-credentials-secret
    fi

    helmTemplateAndApply templates/presto/presto-catalog-config-secret.yaml presto-catalog-config-secret true false
    helmTemplateAndApply templates/presto/presto-common-config.yaml presto-common-config true false
    helmTemplateAndApply templates/presto/presto-coordinator-config.yaml presto-coordinator-config true false
    helmTemplateAndApply templates/presto/presto-serviceaccount.yaml presto-serviceaccount true false
    helmTemplateAndApply templates/presto/presto-jmx-config.yaml presto-jmx-config true false
    helmTemplateAndApply templates/presto/presto-service.yaml presto-service true false
    helmTemplateAndApply templates/presto/presto-worker-config.yaml presto-worker-config true false

    helmTemplateAndApply templates/presto/presto-coordinator-deployment.yaml presto-coordinator-deployment true false
    helmTemplateAndApply templates/presto/presto-worker-deployment.yaml presto-worker-deployment true false
}

installReportingOperatorResources() {
    if [ "$CREATE_REPORTING_OPERATOR_AUTH_PROXY_AUTHENTICATED_EMAILS_SECRET" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-auth-proxy-authenticated-emails-secret.yaml reporting-operator-auth-proxy-authenticated-emails-secret true false
    else
        kubectlDeleteByPruneLabelValue secret reporting-operator-auth-proxy-authenticated-emails-secret
    fi
    if [ "$CREATE_REPORTING_OPERATOR_AUTH_PROXY_COOKIE_SECRET" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-auth-proxy-cookie-secret.yaml reporting-operator-auth-proxy-cookie-secret true false
    else
        kubectlDeleteByPruneLabelValue secret reporting-operator-auth-proxy-cookie-secret
    fi
    if [ "$CREATE_REPORTING_OPERATOR_AUTH_PROXY_HTPASSWD_SECRET" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-auth-proxy-htpasswd-secret.yaml reporting-operator-auth-proxy-htpasswd-secret true false
    else
        kubectlDeleteByPruneLabelValue secret reporting-operator-auth-proxy-htpasswd-secret
    fi
    if [ "$CREATE_REPORTING_OPERATOR_AUTH_PROXY_RBAC" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-auth-proxy-rbac.yaml reporting-operator-auth-proxy-rbac true false
    else
        kubectlDeleteByPruneLabelValue role,rolebinding reporting-operator-auth-proxy-rbac
    fi

    if [ "$CREATE_REPORTING_OPERATOR_PROMETHEUS_BEARER_TOKEN" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-prometheus-bearer-token-secrets.yaml reporting-operator-prometheus-bearer-token-secrets true false
    else
        kubectlDeleteByPruneLabelValue secret reporting-operator-prometheus-bearer-token-secrets
    fi

    if [ "$CREATE_REPORTING_OPERATOR_PROMETHEUS_CERTIFICATE_AUTHORITY" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-prometheus-certificate-authority-config.yaml reporting-operator-prometheus-certificate-authority-config true false
    else
        kubectlDeleteByPruneLabelValue configmap reporting-operator-prometheus-certificate-authority-config
    fi

    if [ "$CREATE_REPORTING_OPERATOR_AWS_CREDENTIALS" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-aws-credentials-secrets.yaml reporting-operator-aws-credentials-secrets true false
    else
        kubectlDeleteByPruneLabelValue role,rolebinding reporting-operator-aws-credentials-secrets
    fi

    helmTemplateAndApply templates/reporting-operator/reporting-operator-rbac.yaml reporting-operator-rbac true false

    if [ "$CREATE_REPORTING_OPERATOR_CLUSTER_MONITORING_VIEW_RBAC" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-cluster-monitoring-view-rbac.yaml reporting-operator-cluster-monitoring-view-rbac true false
    else
        kubectlDeleteByPruneLabelValue clusterrole,clusterrolebinding reporting-operator-cluster-monitoring-view-rbac
    fi

    helmTemplateAndApply templates/reporting-operator/reporting-operator-config.yaml reporting-operator-config true false
    helmTemplateAndApply templates/reporting-operator/reporting-operator-service.yaml reporting-operator-service true false
    helmTemplateAndApply templates/reporting-operator/reporting-operator-serviceaccount.yaml reporting-operator-serviceaccount true false

    if [ "$CREATE_REPORTING_OPERATOR_TLS_SECRETS" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-tls-secrets.yaml reporting-operator-tls-secrets true false
    else
        kubectlDeleteByPruneLabelValue secret reporting-operator-tls-secrets
    fi

    KUBECTL_BIN_OLD="$KUBECTL_BIN"
    KUBECTL_BIN="$OC_BIN"
    if [ "$CREATE_REPORTING_OPERATOR_ROUTE" == "true" ]; then
        helmTemplateAndApply templates/reporting-operator/reporting-operator-route.yaml reporting-operator-route true false
    else
        kubectlDeleteByPruneLabelValue route reporting-operator-route
    fi
    KUBECTL_BIN="$KUBECTL_BIN_OLD"

    helmTemplateAndApply templates/reporting-operator/reporting-operator-deployment.yaml reporting-operator-deployment true false
}

installMeteringResources
installReportingResources
installHdfsResources
installPrestoResources
installReportingOperatorResources

wait

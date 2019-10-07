#! /bin/bash

TEST_OUTPUT_PATH="$(mktemp -d)"

LOG_DIR=$TEST_OUTPUT_PATH/logs
TEST_OUTPUT_DIR=$TEST_OUTPUT_PATH/tests
REPORT_RESULTS_DIR=$TEST_OUTPUT_PATH/report_results
METERINGCONFIGS_DIR=$TEST_OUTPUT_PATH/meteringconfigs
REPORTS_DIR=$TEST_OUTPUT_PATH/reports
DATASOURCES_DIR=$TEST_OUTPUT_PATH/reportdatasources
REPORTQUERIES_DIR=$TEST_OUTPUT_PATH/reportqueries
HIVETABLES_DIR=$TEST_OUTPUT_PATH/hivetables
PRESTOTABLES_DIR=$TEST_OUTPUT_PATH/prestotables
STORAGELOCATIONS_DIR=$TEST_OUTPUT_PATH/storagelocations

TEST_RESULT_REPORT_OUTPUT_DIRECTOR="$TEST_OUTPUT_PATH/reports"

mkdir -p "$LOG_DIR" "$TEST_OUTPUT_DIR" "$REPORT_RESULTS_DIR" "$METERINGCONFIGS_DIR" "$REPORTS_DIR" "$DATASOURCES_DIR" "$REPORTQUERIES_DIR" "$HIVETABLES_DIR" "$PRESTOTABLES_DIR" "$STORAGELOCATIONS_DIR"

echo "Namespace passed: $METERING_TEST_NAMESPACE"

if [[ -z "$METERING_TEST_NAMESPACE" ]]; then
    echo "You need to set \$METERING_TEST_NAMESPACE"
    exit 1
fi

echo "Performing cleanup"

echo "Storing pod descriptions and logs at $LOG_DIR"
echo "Capturing pod descriptions"
PODS="$(kubectl get pods --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r pod; do
    if [[ -n "$pod" ]]; then
        echo "Capturing pod $pod description"
        if ! kubectl describe pod --namespace "$METERING_TEST_NAMESPACE" "$pod" > "$LOG_DIR/${pod}-description.txt"; then
            echo "Error capturing pod $pod description"
        fi
    fi
done <<< "$PODS"

echo "Capturing pod logs"
while read -r pod; do
    if [[ -z "$pod" ]]; then
        continue
    fi
    # There can be multiple containers within a pod. We need to iterate
    # over each of those
    containers=$(kubectl get pods -o jsonpath="{.spec.containers[*].name}" --namespace "$METERING_TEST_NAMESPACE" "$pod")
    for container in $containers; do
        echo "Capturing pod $pod container $container logs"
        if ! kubectl logs --namespace "$METERING_TEST_NAMESPACE" -c "$container" "$pod" > "$LOG_DIR/${pod}-${container}.log"; then
            echo "Error capturing pod $pod container $container logs"
        fi
    done
done <<< "$PODS"

echo "Capturing MeteringConfigs"
METERINGCONFIGS="$(kubectl get meteringconfigs --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r meteringconfig; do
    if [[ -n "$meteringconfig" ]]; then
        echo "Capturing MeteringConfig $meteringconfig as json"
        if ! kubectl get meteringconfig "$meteringconfig" --namespace "$METERING_TEST_NAMESPACE" -o json > "$METERINGCONFIGS_DIR/${meteringconfig}.json"; then
            echo "Error getting $meteringconfig as json"
        fi
    fi
done <<< "$METERINGCONFIGS"

echo "Capturing Metering StorageLocations"
STORAGELOCATIONS="$(kubectl get storagelocations --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r storagelocation; do
    if [[ -n "$storagelocation" ]]; then
        echo "Capturing StorageLocation $storagelocation as json"
        if ! kubectl get storagelocation "$storagelocation" --namespace "$METERING_TEST_NAMESPACE" -o json > "$STORAGELOCATIONS_DIR/${storagelocation}.json"; then
            echo "Error getting $storagelocation as json"
        fi
    fi
done <<< "$STORAGELOCATIONS"

echo "Capturing Metering PrestoTables"
PRESTOTABLES="$(kubectl get prestotables --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r prestotable; do
    if [[ -n "$prestotable" ]]; then
        echo "Capturing PrestoTable $prestotable as json"
        if ! kubectl get prestotable "$prestotable" --namespace "$METERING_TEST_NAMESPACE" -o json > "$PRESTOTABLES_DIR/${prestotable}.json"; then
            echo "Error getting $prestotable as json"
        fi
    fi
done <<< "$PRESTOTABLES"

echo "Capturing Metering HiveTables"
HIVETABLES="$(kubectl get hivetables --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r hivetable; do
    if [[ -n "$hivetable" ]]; then
        echo "Capturing HiveTable $hivetable as json"
        if ! kubectl get hivetable "$hivetable" --namespace "$METERING_TEST_NAMESPACE" -o json > "$HIVETABLES_DIR/${hivetable}.json"; then
            echo "Error getting $hivetable as json"
        fi
    fi
done <<< "$HIVETABLES"

echo "Capturing Metering ReportDataSources"
DATASOURCES="$(kubectl get reportdatasources --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r datasource; do
    if [[ -n "$datasource" ]]; then
        echo "Capturing ReportDataSource $datasource as json"
        if ! kubectl get reportdatasource "$datasource" --namespace "$METERING_TEST_NAMESPACE" -o json > "$DATASOURCES_DIR/${datasource}.json"; then
            echo "Error getting $datasource as json"
        fi
    fi
done <<< "$DATASOURCES"

echo "Capturing Metering ReportQueries"
RGQS="$(kubectl get reportqueries --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r rgq; do
    if [[ -n "$rgq" ]]; then
        echo "Capturing ReportQuery $rgq as json"
        if ! kubectl get reportquery "$rgq" --namespace "$METERING_TEST_NAMESPACE" -o json > "$REPORTQUERIES_DIR/${rgq}.json"; then
            echo "Error getting $rgq as json"
        fi
    fi
done <<< "$RGQS"

echo "Capturing Metering Reports"
REPORTS="$(kubectl get reports --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r report; do
    if [[ -n "$report" ]]; then
        echo "Capturing Report $report as json"
        if ! kubectl get report "$report" --namespace "$METERING_TEST_NAMESPACE" -o json > "$REPORTS_DIR/${report}.json"; then
            echo "Error getting $report as json"
        fi
    fi
done <<< "$REPORTS"

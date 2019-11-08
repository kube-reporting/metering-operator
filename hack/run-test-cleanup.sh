***REMOVED***

echo "Performing cleanup"
echo "Storing pod descriptions and logs at $LOG_DIR"
echo "Capturing pod descriptions"

PODS="$(kubectl get pods --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r pod; do
    if [[ -n "$pod" ]]; then
        echo "Capturing pod $pod description"
        if ! kubectl describe pod --namespace "$METERING_TEST_NAMESPACE" "$pod" > "$LOG_DIR/${pod}-description.txt"; then
            echo "Error capturing pod $pod description"
        ***REMOVED***
    ***REMOVED***
done <<< "$PODS"

echo "Capturing pod logs"
while read -r pod; do
    if [[ -z "$pod" ]]; then
        continue
    ***REMOVED***
    # There can be multiple containers within a pod. We need to iterate
    # over each of those
    containers=$(kubectl get pods -o jsonpath="{.spec.containers[*].name}" --namespace "$METERING_TEST_NAMESPACE" "$pod")
    for container in $containers; do
        echo "Capturing pod $pod container $container logs"
        if ! kubectl logs --namespace "$METERING_TEST_NAMESPACE" -c "$container" "$pod" > "$LOG_DIR/${pod}-${container}.log"; then
            echo "Error capturing pod $pod container $container logs"
        ***REMOVED***
    done
done <<< "$PODS"

echo "Capturing MeteringCon***REMOVED***gs"
METERINGCONFIGS="$(kubectl get meteringcon***REMOVED***gs --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r meteringcon***REMOVED***g; do
    if [[ -n "$meteringcon***REMOVED***g" ]]; then
        echo "Capturing MeteringCon***REMOVED***g $meteringcon***REMOVED***g as json"
        if ! kubectl get meteringcon***REMOVED***g "$meteringcon***REMOVED***g" --namespace "$METERING_TEST_NAMESPACE" -o json > "$METERINGCONFIGS_DIR/${meteringcon***REMOVED***g}.json"; then
            echo "Error getting $meteringcon***REMOVED***g as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$METERINGCONFIGS"

echo "Capturing Metering StorageLocations"
STORAGELOCATIONS="$(kubectl get storagelocations --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r storagelocation; do
    if [[ -n "$storagelocation" ]]; then
        echo "Capturing StorageLocation $storagelocation as json"
        if ! kubectl get storagelocation "$storagelocation" --namespace "$METERING_TEST_NAMESPACE" -o json > "$STORAGELOCATIONS_DIR/${storagelocation}.json"; then
            echo "Error getting $storagelocation as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$STORAGELOCATIONS"

echo "Capturing Metering PrestoTables"
PRESTOTABLES="$(kubectl get prestotables --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r prestotable; do
    if [[ -n "$prestotable" ]]; then
        echo "Capturing PrestoTable $prestotable as json"
        if ! kubectl get prestotable "$prestotable" --namespace "$METERING_TEST_NAMESPACE" -o json > "$PRESTOTABLES_DIR/${prestotable}.json"; then
            echo "Error getting $prestotable as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$PRESTOTABLES"

echo "Capturing Metering HiveTables"
HIVETABLES="$(kubectl get hivetables --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r hivetable; do
    if [[ -n "$hivetable" ]]; then
        echo "Capturing HiveTable $hivetable as json"
        if ! kubectl get hivetable "$hivetable" --namespace "$METERING_TEST_NAMESPACE" -o json > "$HIVETABLES_DIR/${hivetable}.json"; then
            echo "Error getting $hivetable as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$HIVETABLES"

echo "Capturing Metering ReportDataSources"
DATASOURCES="$(kubectl get reportdatasources --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r datasource; do
    if [[ -n "$datasource" ]]; then
        echo "Capturing ReportDataSource $datasource as json"
        if ! kubectl get reportdatasource "$datasource" --namespace "$METERING_TEST_NAMESPACE" -o json > "$DATASOURCES_DIR/${datasource}.json"; then
            echo "Error getting $datasource as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$DATASOURCES"

echo "Capturing Metering ReportQueries"
RGQS="$(kubectl get reportqueries --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r rgq; do
    if [[ -n "$rgq" ]]; then
        echo "Capturing ReportQuery $rgq as json"
        if ! kubectl get reportquery "$rgq" --namespace "$METERING_TEST_NAMESPACE" -o json > "$REPORTQUERIES_DIR/${rgq}.json"; then
            echo "Error getting $rgq as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$RGQS"

echo "Capturing Metering Reports"
REPORTS="$(kubectl get reports --no-headers --namespace "$METERING_TEST_NAMESPACE" -o name | cut -d/ -f2)"
while read -r report; do
    if [[ -n "$report" ]]; then
        echo "Capturing Report $report as json"
        if ! kubectl get report "$report" --namespace "$METERING_TEST_NAMESPACE" -o json > "$REPORTS_DIR/${report}.json"; then
            echo "Error getting $report as json"
        ***REMOVED***
    ***REMOVED***
done <<< "$REPORTS"

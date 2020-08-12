#! /bin/bash

# TODO: adapt to make this more general for debugging support cases
METERING_TEST_NAMESPACE="${METERING_TEST_NAMESPACE:=$METERING_NAMESPACE}"
METERING_GATHER_OLM_RESOURCES="${METERING_GATHER_OLM_RESOURCES:=true}"

LOG_DIR="${LOG_DIR:=$PWD/must-gather}"
POD_LOG_PATH=${POD_LOG_PATH:="${LOG_DIR}/pod_logs"}
mkdir -p ${POD_LOG_PATH}/

resources=()
resources+=(pods)
resources+=(deployments)
resources+=(statefulsets)
resources+=(services)
resources+=(hivetables)
resources+=(prestotables)
resources+=(storagelocations)
resources+=(meteringconfigs)
resources+=(reportdatasources)
resources+=(reportqueries)
resources+=(reports)

if [[ ${METERING_GATHER_OLM_RESOURCES} == "true" ]]; then
    resources+=(subscriptions)
    resources+=(operatorgroups)
    resources+=(clusterserviceversions)
    resources+=(installplans)
fi

echo "Storing the must-gather output at $LOG_DIR"
for resource in "${resources[@]}"; do
    echo "Collecting dump of ${resource} in the ${METERING_TEST_NAMESPACE} namespace" | tee -a  ${LOG_DIR}/gather-debug.log
    { timeout 120 oc adm --namespace ${METERING_TEST_NAMESPACE} --dest-dir=${LOG_DIR} inspect "${resource}"; } >> ${LOG_DIR}/gather-debug.log 2>&1
done

echo "Collecting the metering-related CRDs from the cluster"
for resource in $(oc get crd | grep metering | awk '{ print $1 }'); do
    timeout 120 oc adm --dest-dir=${LOG_DIR} inspect "crd/$resource" >> ${LOG_DIR}/gather-debug.log 2>&1
done

commands=()
commands+=("get pods -o wide")
commands+=("get reportdatasources")
commands+=("get reports")
commands+=("get prestotables")
commands+=("get events")

for command in "${commands[@]}"; do
     echo "Collecting output of the following oc command: 'oc ${command}'" | tee -a ${LOG_DIR}/gather-debug.log
     COMMAND_OUTPUT_FILE=${POD_LOG_PATH}/${command// /_}
     timeout 120 oc -n ${METERING_TEST_NAMESPACE} ${command} >> "${COMMAND_OUTPUT_FILE}"
done

for pod in $(oc --namespace $METERING_TEST_NAMESPACE get pods --no-headers -o name | cut -d/ -f2); do
    for container in $(oc --namespace $METERING_TEST_NAMESPACE get pods -o jsonpath="{.spec.containers[*].name}" "$pod"); do
        echo "Capturing Pod $pod container $container logs"
        if ! oc logs --namespace "$METERING_TEST_NAMESPACE" -c "$container" "$pod" >> "${POD_LOG_PATH}/${pod}-${container}.log"; then
            echo "Error capturing pod $pod container $container logs"
        fi
    done
done

echo "Deleting any empty test artifact files" >> ${LOG_DIR}/gather-debug.log
find "${LOG_DIR}" -empty -delete >> ${LOG_DIR}/gather-debug.log 2>&1
exit 0

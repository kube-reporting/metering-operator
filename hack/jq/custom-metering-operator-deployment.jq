def selectDeploymentEnv(deployment; key):
    (deployment | .spec.template.spec.containers[].env[] | select(.name == key))
;
def updateDeploymentEnv(deployment; key; value):
    selectDeploymentEnv(deployment; key) = {name: key, value: value}
;
def isEmpty(v): v == null or (v | length) == 0;
def notEmpty(v): isEmpty(v) | not;

.

| if notEmpty($ENV.METERING_OPERATOR_IMAGE_REPO) and notEmpty($ENV.METERING_OPERATOR_IMAGE_TAG) then
    # Update the image for each container in the pod
    .spec.template.spec.containers[].image = ($ENV.METERING_OPERATOR_IMAGE_REPO + ":" + $ENV.METERING_OPERATOR_IMAGE_TAG)
***REMOVED***
    .
end

| if notEmpty($ENV.METERING_OPERATOR_ALL_NAMESPACES) then
    # Update the env var ALL_NAMESPACES
     updateDeploymentEnv(.; "ALL_NAMESPACES"; $ENV.METERING_OPERATOR_ALL_NAMESPACES)
***REMOVED***
    .
end
| if notEmpty($ENV.METERING_OPERATOR_TARGET_NAMESPACES) then
    # Update the env var TARGET_NAMESPACES
    updateDeploymentEnv(.; "TARGET_NAMESPACES"; $ENV.METERING_OPERATOR_TARGET_NAMESPACES)
***REMOVED***
    .
end

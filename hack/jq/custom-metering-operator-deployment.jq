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
else
    .
end

| if notEmpty($ENV.METERING_OPERATOR_TARGET_NAMESPACES) then
    # Update the env var WATCH_NAMESPACE
    updateDeploymentEnv(.; "WATCH_NAMESPACE"; $ENV.METERING_OPERATOR_TARGET_NAMESPACES)
else
    .
end

#!/bin/bash

# lowercase the value, and characters we use in branches with dashes
function sanetize_namespace() {
    echo -n "$1" | tr '[:upper:]' '[:lower:]' | tr '.' '-' | sed 's/[._]/-/g'
}

function kubectl_cmd() {
    kubectl --namespace="${METERING_NAMESPACE}" "$@"
}

function kube-install() {
  local files
  IFS=" " read -r -a files <<< "$(kubectl_files "$@")"
  kubectl_cmd apply "${files[@]}"
}

function kube-remove-non-file() {
  kubectl_cmd delete "$@"
}

function kube-remove() {
  IFS=" " read -r -a files <<< "$(kubectl_files "$@")"
  kubectl_cmd delete "${files[@]}"
}

function msg() {
  echo -e "\x1b[1;35m${@}\x1b[0m"
}

# formats flags for kubectl for the given files
function kubectl_files() {
  local files=()
  for f in "${@}"; do
      files+=(-f "$f")
  done
  echo "${files[@]}"
}

function install_metering() {
    INSTALL_METHOD=$1
    echo "Installing metering using "$INSTALL_METHOD" install method"
    if [[ "$INSTALL_METHOD" == "direct" || "$INSTALL_METHOD" == "generic-direct" ]]; then
        "$ROOT_DIR/hack/install.sh"
    elif [ "$INSTALL_METHOD" == "openshift-direct" ]; then
        "$ROOT_DIR/hack/openshift-install.sh"
    elif [ "$INSTALL_METHOD" == "olm" ]; then
        "$ROOT_DIR/hack/olm-install.sh"
    else
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    fi
}

function uninstall_metering() {
    INSTALL_METHOD=$1
    echo "Uninstalling metering using "$INSTALL_METHOD" uninstall method"
    if [[ "$INSTALL_METHOD" == "direct" || "$INSTALL_METHOD" == "generic-direct" ]]; then
        "$ROOT_DIR/hack/uninstall.sh"
    elif [ "$INSTALL_METHOD" == "openshift-direct" ]; then
        "$ROOT_DIR/hack/openshift-uninstall.sh"
    elif [ "$INSTALL_METHOD" == "olm" ]; then
        "$ROOT_DIR/hack/olm-uninstall.sh"
    else
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    fi
}

# Taken and modified slightly from https://github.com/kubernetes/charts/blob/f1711c220988b69e530263dc924eaed0a759e441/test/changed.sh#L42
capture_pod_logs() {
    NS="$(sanetize_namespace "$1")"
    echo "Capturing logs for $NS"
    # List all logs for all containers in all pods for the namespace which was
    PODS="$(kubectl get pods --no-headers --namespace "$NS")"
    echo '===Pods==='
    echo "$PODS"
    echo "$PODS" | awk '{ print $1 }' | while read -r pod; do
        if [[ -n "$pod" ]]; then
            printf '===Details from pod %s:===\n' "$pod"

            printf '...Description of pod %s:...\n' "$pod"
            kubectl describe pod --namespace "$NS" "$pod" || true
            printf '...End of description for pod %s...\n\n' "$pod"

            # There can be multiple containers within a pod. We need to iterate
            # over each of those
            containers=$(kubectl get pods -o jsonpath="{.spec.containers[*].name}" --namespace "$NS" "$pod")
            for container in $containers; do
                printf -- '---Logs from container %s in pod %s:---\n' "$container" "$pod"
                kubectl logs --namespace "$METERING_NAMESPACE" -c "$container" "$pod" || true
                printf -- '---End of logs for container %s in pod %s---\n\n' "$container" "$pod"
            done

            printf '===End of details for pod %s===\n\n' "$pod"
        fi
    done
}

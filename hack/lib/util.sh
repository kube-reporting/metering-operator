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
    elif [ "$INSTALL_METHOD" == "tectonic-direct" ]; then
        "$ROOT_DIR/hack/tectonic-install.sh"
    elif [ "$INSTALL_METHOD" == "openshift-direct" ]; then
        "$ROOT_DIR/hack/openshift-install.sh"
    elif [ "$INSTALL_METHOD" == "alm" ]; then
        "$ROOT_DIR/hack/alm-install.sh"
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
    elif [ "$INSTALL_METHOD" == "tectonic-direct" ]; then
        "$ROOT_DIR/hack/tectonic-uninstall.sh"
    elif [ "$INSTALL_METHOD" == "openshift-direct" ]; then
        "$ROOT_DIR/hack/openshift-uninstall.sh"
    elif [ "$INSTALL_METHOD" == "alm" ]; then
        "$ROOT_DIR/hack/alm-uninstall.sh"
    else
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    fi
}

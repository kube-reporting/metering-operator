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
    echo "Installing metering using the "$INSTALL_METHOD" install method"
    if [[ "$INSTALL_METHOD" == "openshift" ]]; then
        "$ROOT_DIR/bin/deploy-metering" install --log-level debug
    # TODO: update the deploy package/cli to handle OLM
    elif [ "$INSTALL_METHOD" == "olm" ]; then
        "$ROOT_DIR/hack/olm-install.sh"
    else
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    fi
}

function uninstall_metering() {
    INSTALL_METHOD=$1
    echo "Uninstalling metering using the "$INSTALL_METHOD" uninstall method"
    if [[ "$INSTALL_METHOD" == "openshift" ]]; then
        "$ROOT_DIR/bin/deploy-metering" uninstall --log-level debug
    # TODO: update the deploy package/cli to handle OLM
    elif [ "$INSTALL_METHOD" == "olm" ]; then
        "$ROOT_DIR/hack/olm-uninstall.sh"
    else
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    fi
}


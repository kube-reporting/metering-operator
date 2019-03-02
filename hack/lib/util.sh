#!/bin/bash

# lowercase the value, and characters we use in branches with dashes
function sanetize_namespace() {
    echo -n "$1" | tr '[:upper:]' '[:lower:]' | tr '.' '-' | sed 's/[._]/-/g'
}

function kubectl_cmd() {
    kubectl --namespace="${METERING_NAMESPACE}" "$@"
}

function kube-install() {
  local ***REMOVED***les
  IFS=" " read -r -a ***REMOVED***les <<< "$(kubectl_***REMOVED***les "$@")"
  kubectl_cmd apply "${***REMOVED***les[@]}"
}

function kube-remove-non-***REMOVED***le() {
  kubectl_cmd delete "$@"
}

function kube-remove() {
  IFS=" " read -r -a ***REMOVED***les <<< "$(kubectl_***REMOVED***les "$@")"
  kubectl_cmd delete "${***REMOVED***les[@]}"
}

function msg() {
  echo -e "\x1b[1;35m${@}\x1b[0m"
}

# formats flags for kubectl for the given ***REMOVED***les
function kubectl_***REMOVED***les() {
  local ***REMOVED***les=()
  for f in "${@}"; do
      ***REMOVED***les+=(-f "$f")
  done
  echo "${***REMOVED***les[@]}"
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
    ***REMOVED***
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    ***REMOVED***
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
    ***REMOVED***
        echo "Invalid \$INSTALL_METHOD: $INSTALL_METHOD"
        exit 1
    ***REMOVED***
}


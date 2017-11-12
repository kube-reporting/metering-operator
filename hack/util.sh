# This should not be invoked directly. It provides functions and data for other scripts.

: "${CHARGEBACK_NAMESPACE:=team-chargeback}"
: "${PULL_SECRET_NAMESPACE:=tectonic-system}"
: "${PULL_SECRET:=coreos-pull-secret}"

function kubectl_cmd() {
    echo "kubectl --namespace=${CHARGEBACK_NAMESPACE}"
}

function kube-install() {
  local cmd=$(kubectl_cmd)
  local ***REMOVED***les=$(kubectl_***REMOVED***les $@)
  ${cmd} apply ${***REMOVED***les}
}

function kube-remove-non-***REMOVED***le() {
  local cmd=$(kubectl_cmd)
  ${cmd} delete $@
}

function kube-remove() {
  local cmd=$(kubectl_cmd)
  local ***REMOVED***les=$(kubectl_***REMOVED***les $@)
  ${cmd} delete ${***REMOVED***les}
}

function msg() {
  echo -e "\x1b[1;35m${@}\x1b[0m"
}

function copy-tectonic-pull() {
  local pullSecret=$(kubectl --namespace=${PULL_SECRET_NAMESPACE} get secrets ${PULL_SECRET} -o json --export)
  pullSecret="${pullSecret/${PULL_SECRET_NAMESPACE}/${CHARGEBACK_NAMESPACE}}"
  echo ${pullSecret} | kube-install -
}

# formats flags for kubectl for the given ***REMOVED***les
function kubectl_***REMOVED***les() {
  local str=""
  for f in "${@}"; do
    str="${str-} -f ${f}"
  done
  echo ${str}
}

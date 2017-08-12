# This should not be invoked directly. It provides functions and data for other scripts.

CHARGEBACK_NAMESPACE="tectonic-chargeback"
TECTONIC_NAMESPACE="tectonic-system"
TECTONIC_PULL_SECRET="coreos-pull-secret"
AWS_SECRET="aws"

function kube-install() {
  local ***REMOVED***les=$(kubectl_***REMOVED***les $@)
  kubectl apply ${***REMOVED***les}
}

function kube-remove() {
  local ***REMOVED***les=$(kubectl_***REMOVED***les $@)
  kubectl delete ${***REMOVED***les}
}

function msg() {
  echo -e "\x1b[1;35m${@}\x1b[0m"
}

function copy-tectonic-pull() {
  local pullSecret=$(kubectl --namespace=${TECTONIC_NAMESPACE} get secrets ${TECTONIC_PULL_SECRET} -o json)
  pullSecret="${pullSecret/${TECTONIC_NAMESPACE}/${CHARGEBACK_NAMESPACE}}"
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

function aws_secret() {
  local id=${1}
  local secret=${2}
  cat <<EOF
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "${AWS_SECRET}",
        "namespace": "${CHARGEBACK_NAMESPACE}"
    },
    "data": {
        "AWS_ACCESS_KEY_ID": "${id}",
        "AWS_SECRET_ACCESS_KEY": "${secret}"
    },
    "type": "Opaque"
}
EOF
}

# This should not be invoked directly. It provides functions and data for other scripts.

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

# formats flags for kubectl for the given ***REMOVED***les
function kubectl_***REMOVED***les() {
  local str=""
  for f in "${@}"; do
    str="${str-} -f ${f}"
  done
  echo ${str}
}

# This should not be invoked directly. It provides functions and data for other scripts.

function kube-install() {
  local files=$(kubectl_files $@)
  kubectl apply ${files}
}

function kube-remove() {
  local files=$(kubectl_files $@)
  kubectl delete ${files}
}

function msg() {
  echo -e "\x1b[1;35m${@}\x1b[0m"
}

# formats flags for kubectl for the given files
function kubectl_files() {
  local str=""
  for f in "${@}"; do
    str="${str-} -f ${f}"
  done
  echo ${str}
}

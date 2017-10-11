#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this ***REMOVED***le except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the speci***REMOVED***c language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_PACKAGE=github.com/coreos-inc/kube-chargeback
SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
SCRIPT_BASE=${SCRIPT_ROOT}
# SCRIPT_BASE=${SCRIPT_ROOT}/../..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo k8s.io/code-generator)}

clientgen="${PWD}/client-gen-binary"
listergen="${PWD}/lister-gen"
informergen="${PWD}/informer-gen"
deepcopygen="${PWD}/deepcopy-gen"
# Register function to be called on EXIT to remove generated binary.
function cleanup {
  rm -f "${clientgen:-}"
  rm -f "${listergen:-}"
  rm -f "${informergen:-}"
  rm -f "${deepcopygen:-}"
}
trap cleanup EXIT

function generate_group() {
  local GROUP_NAME=$1
  local VERSION=$2

  local APIS_PKG=${SCRIPT_PACKAGE}/pkg/apis
  local CLIENT_PKG=${SCRIPT_PACKAGE}/pkg/generated
  local CLIENTSET_PKG=${CLIENT_PKG}/clientset
  local LISTERS_PKG=${CLIENT_PKG}/listers
  local INFORMERS_PKG=${CLIENT_PKG}/informers

  local INPUT_APIS=(
    ${GROUP_NAME}/
    ${GROUP_NAME}/${VERSION}
  )


  echo "Building deepcopy-gen"
  go build -o "${deepcopygen}" ${CODEGEN_PKG}/cmd/deepcopy-gen

  echo "generating deepcopy funcs for group ${GROUP_NAME} and version ${VERSION}"
  ${deepcopygen} --input-dirs ${APIS_PKG}/${GROUP_NAME}/${VERSION} -O zz_generated.deepcopy --go-header-***REMOVED***le $SCRIPT_ROOT/vendor/k8s.io/gengo/boilerplate/no-boilerplate.go.txt

  echo "Building client-gen"
  go build -o "${clientgen}" ${CODEGEN_PKG}/cmd/client-gen

  echo "generating clientset for group ${GROUP_NAME} and version ${VERSION} at ${CLIENTSET_PKG}"
  ${clientgen} --clientset-name="versioned" --input-base ${APIS_PKG} --input ${GROUP_NAME}/${VERSION} --clientset-path ${CLIENTSET_PKG} --go-header-***REMOVED***le $SCRIPT_ROOT/vendor/k8s.io/gengo/boilerplate/no-boilerplate.go.txt


  echo "Building lister-gen"
  go build -o "${listergen}" ${CODEGEN_PKG}/cmd/lister-gen

  echo "generating listers for group ${GROUP_NAME} and version ${VERSION} at ${LISTERS_PKG}"
  ${listergen} --input-dirs ${APIS_PKG}/${GROUP_NAME} --input-dirs ${APIS_PKG}/${GROUP_NAME}/${VERSION} --output-package ${LISTERS_PKG} --go-header-***REMOVED***le $SCRIPT_ROOT/vendor/k8s.io/gengo/boilerplate/no-boilerplate.go.txt

  echo "Building informer-gen"
  go build -o "${informergen}" ${CODEGEN_PKG}/cmd/informer-gen

  echo "generating informers for group ${GROUP_NAME} and version ${VERSION} at ${INFORMERS_PKG}"
  ${informergen} \
    --input-dirs ${APIS_PKG}/${GROUP_NAME} --input-dirs ${APIS_PKG}/${GROUP_NAME}/${VERSION} \
    --versioned-clientset-package ${CLIENTSET_PKG}/versioned \
    --listers-package ${LISTERS_PKG} \
    --output-package ${INFORMERS_PKG} \
    --go-header-***REMOVED***le $SCRIPT_ROOT/vendor/k8s.io/gengo/boilerplate/no-boilerplate.go.txt
}

generate_group chargeback v1alpha1

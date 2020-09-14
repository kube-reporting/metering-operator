#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_PACKAGE=github.com/kube-reporting/metering-operator
SCRIPT_ROOT="$(realpath $(dirname ${BASH_SOURCE[0]})/..)"
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo k8s.io/code-generator)}

set -x

# generate kubernetes client
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
    "${SCRIPT_PACKAGE}/pkg/generated" "${SCRIPT_PACKAGE}/pkg/apis" \
    metering:v1 \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# generate-groups doesn't do defaulters
echo "Generating defaulters"
${GOPATH}/bin/defaulter-gen \
    --input-dirs "${SCRIPT_PACKAGE}/pkg/apis/metering/v1" \
    -O zz_generated.defaults \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# generate mocks
go build -v -o "$SCRIPT_ROOT/vendor/mockgen" "./vendor/github.com/golang/mock/mockgen"
"$SCRIPT_ROOT/vendor/mockgen" \
    -package mockprestostore \
    -destination "$SCRIPT_ROOT/pkg/operator/prestostore/mock/reports.go" \
    "$SCRIPT_PACKAGE/pkg/operator/prestostore" \
    ReportResultsRepo
gofmt -w "$SCRIPT_ROOT/pkg/operator/prestostore/mock/reports.go"

SHELL := /bin/bash

ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
include build/check_defined.mk

# Package
GO_PKG := github.com/operator-framework/operator-metering
REPORTING_OPERATOR_PKG := $(GO_PKG)/cmd/reporting-operator

DOCKER_BUILD_ARGS ?=

DOCKER_BUILD_CMD = docker build $(DOCKER_BUILD_ARGS)

USE_IMAGEBUILDER ?= false

ifeq ($(USE_IMAGEBUILDER), true)
	DOCKER_BUILD_CMD = imagebuilder
endif

GO_BUILD_ARGS := -ldflags '-extldflags "-static"'
GOOS = "linux"
CGO_ENABLED = 0

REPORTING_OPERATOR_BIN_OUT = bin/reporting-operator
REPORTING_OPERATOR_BIN_OUT_LOCAL = bin/reporting-operator-local
RUN_UPDATE_CODEGEN ?= true
CHECK_GO_FILES ?= true

REPORTING_OPERATOR_BIN_DEPENDENCIES =
CODEGEN_SOURCE_GO_FILES =
CODEGEN_OUTPUT_GO_FILES =
REPORTING_OPERATOR_GO_FILES =

# Adds all the Go files in the repo as a dependency to the build-reporting-operator target
ifeq ($(CHECK_GO_FILES), true)
	JQ_DEP_SCRIPT = '.Deps[] | select(. | contains("$(GO_PKG)"))'
	REPORTING_OPERATOR_GO_FILES = $(shell go list -json $(REPORTING_OPERATOR_PKG) | jq $(JQ_DEP_SCRIPT) -r | xargs -I{} find $(GOPATH)/src/$(REPORTING_OPERATOR_PKG) $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)
endif

# Adds the update-codegen dependency to the build-reporting-operator target
ifeq ($(RUN_UPDATE_CODEGEN), true)
	REPORTING_OPERATOR_BIN_DEPENDENCIES += update-codegen
	CODEGEN_SOURCE_GO_FILES = $(shell $(ROOT_DIR)/hack/codegen_source_files.sh)
	CODEGEN_OUTPUT_GO_FILES = $(shell $(ROOT_DIR)/hack/codegen_output_files.sh)
endif


DOCKER_COMMON_NAMES := \
	reporting-operator \
	helm-operator \
	metering-operator \
	hadoop \
	hive \
	metering-e2e

DOCKER_BUILD_NAMES = $(DOCKER_COMMON_NAMES)
DOCKER_TAG_NAMES = $(DOCKER_COMMON_NAMES)
DOCKER_PUSH_NAMES = $(DOCKER_COMMON_NAMES)

DOCKER_BASE_URL := quay.io/coreos

METERING_OPERATOR_IMAGE := $(DOCKER_BASE_URL)/metering-helm-operator
REPORTING_OPERATOR_IMAGE := $(DOCKER_BASE_URL)/metering-reporting-operator
HELM_OPERATOR_IMAGE := $(DOCKER_BASE_URL)/helm-operator
HADOOP_IMAGE := $(DOCKER_BASE_URL)/metering-hadoop
HIVE_IMAGE := $(DOCKER_BASE_URL)/metering-hive
METERING_E2E_IMAGE := $(DOCKER_BASE_URL)/metering-e2e

GIT_SHA    := $(shell git rev-parse HEAD)
GIT_TAG    := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
RELEASE_TAG := $(shell hack/print-version.sh)

PULL_TAG_IMAGE_SOURCE ?= false
USE_LATEST_TAG ?= false
USE_RELEASE_TAG = true
PUSH_RELEASE_TAG = false

DOCKER_BUILD_CONTEXT = $(dir $(DOCKERFILE))
IMAGE_TAG = $(GIT_SHA)
TAG_IMAGE_SOURCE = $(IMAGE_NAME):$(GIT_SHA)

# Hive Git repository for Thrift definitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

# TODO: Add tests
all: fmt test docker-build-all

# Usage:
#	make docker-build DOCKERFILE= IMAGE_NAME=

docker-build:
	$(DOCKER_BUILD_CMD) -t $(IMAGE_NAME):$(GIT_SHA) -f $(DOCKERFILE) $(DOCKER_BUILD_CONTEXT)
ifdef BRANCH_TAG
	$(MAKE) docker-tag IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(BRANCH_TAG)
endif
ifdef DEPLOY_TAG
	$(MAKE) docker-tag IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(DEPLOY_TAG)
endif
ifneq ($(GIT_TAG),)
	$(MAKE) docker-tag IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(GIT_TAG)
endif
ifeq ($(USE_RELEASE_TAG), true)
	$(MAKE) docker-tag IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(RELEASE_TAG)
endif
ifeq ($(USE_LATEST_TAG), true)
	$(MAKE) docker-tag IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=latest
endif

# Usage:
#	make docker-tag SOURCE_IMAGE=$(IMAGE_NAME):$(GIT_SHA) IMAGE_NAME= IMAGE_TAG=
docker-tag:
ifeq ($(PULL_TAG_IMAGE_SOURCE), true)
	$(MAKE) docker-pull IMAGE=$(TAG_IMAGE_SOURCE)
endif
	docker tag $(TAG_IMAGE_SOURCE) $(IMAGE_NAME):$(IMAGE_TAG)

# Usage:
#	make docker-pull IMAGE=

docker-pull:
	docker pull $(IMAGE)

# Usage:
#	make docker-push IMAGE_NAME= IMAGE_TAG=

docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
ifeq ($(PUSH_RELEASE_TAG), true)
	docker push $(IMAGE_NAME):$(RELEASE_TAG)
endif
ifeq ($(USE_LATEST_TAG), true)
	docker push $(IMAGE_NAME):latest
endif
ifneq ($(GIT_TAG),)
	docker push $(IMAGE_NAME):$(GIT_TAG)
endif
ifdef BRANCH_TAG
	docker push $(IMAGE_NAME):$(BRANCH_TAG)
endif
ifdef DEPLOY_TAG
	docker push $(IMAGE_NAME):$(DEPLOY_TAG)
endif

# These generate new make targets like metering-operator-docker-build
# which can be invoked.
DOCKER_BUILD_TARGETS := $(addsuffix -docker-build, $(DOCKER_BUILD_NAMES))
DOCKER_PUSH_TARGETS := $(addsuffix -docker-push, $(DOCKER_PUSH_NAMES))
DOCKER_TAG_TARGETS := $(addsuffix -docker-tag, $(DOCKER_TAG_NAMES))
DOCKER_PULL_TARGETS := $(addsuffix -docker-pull, $(DOCKER_PUSH_NAMES))

# The steps below run for each value of $(DOCKER_TARGETS) effectively, generating multiple Make targets.
# To make it easier to follow, each step will include an example after the evaluation.
# The example will be using the metering-operator targets as it's example.
#
# The pattern/string manipulation below does the following (starting from the inner most expression):
# 1) strips -docker-push, -docker-tag, or -docker-pull from the target name ($@) giving us the non suffixed value from $(TARGETS)
# ex: metering-operator-docker-build -> metering-operator
# 2) Replaces - with _
# ex: metering-operator -> metering_helm_operator
# 3) Uppercases letters
# ex: metering_helm_operator -> METERING_HELM_OPERATOR
# 4) Appends _IMAGE
# ex: METERING_HELM_OPERATOR -> METERING_OPERATOR_IMAGE
# That gives us the value for the docker-build, docker-tag, or docker-push IMAGE_NAME variable.

$(DOCKER_PUSH_TARGETS)::
	$(MAKE) docker-push IMAGE_NAME=$($(addsuffix _IMAGE, $(shell echo $(subst -,_,$(subst -docker-push,,$@)) | tr a-z A-Z)))

$(DOCKER_TAG_TARGETS)::
	$(MAKE) docker-tag IMAGE_NAME=$($(addsuffix _IMAGE, $(shell echo $(subst -,_,$(subst -docker-tag,,$@)) | tr a-z A-Z)))

$(DOCKER_PULL_TARGETS)::
	$(MAKE) docker-pull IMAGE_NAME=$($(addsuffix _IMAGE, $(shell echo $(subst -,_,$(subst -docker-pull,,$@)) | tr a-z A-Z)))

docker-build-all: $(DOCKER_BUILD_TARGETS)

docker-push-all: $(DOCKER_PUSH_TARGETS)

docker-tag-all: $(DOCKER_TAG_TARGETS)

docker-pull-all: $(DOCKER_PULL_TARGETS)

reporting-operator-docker-build: Dockerfile.reporting-operator
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(REPORTING_OPERATOR_IMAGE) DOCKER_BUILD_CONTEXT=$(ROOT_DIR)

metering-e2e-docker-build: Dockerfile.e2e
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(METERING_E2E_IMAGE) DOCKER_BUILD_CONTEXT=$(ROOT_DIR)

metering-operator-docker-build: Dockerfile.metering-operator helm-operator-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(METERING_OPERATOR_IMAGE) DOCKER_BUILD_CONTEXT=$(ROOT_DIR)

helm-operator-docker-build: images/helm-operator/Dockerfile $(find -type f images/helm-operator)
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HELM_OPERATOR_IMAGE) USE_LATEST_TAG=true

hadoop-docker-build: images/hadoop/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE) USE_LATEST_TAG=true

hive-docker-build: images/hive/Dockerfile hadoop-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: Gopkg.toml
	dep ensure -v

test:
	go test ./pkg/...

test-docker:
	docker run -i $(METERING_E2E_IMAGE):$(IMAGE_TAG) bash -c 'make test'

# Runs gofmt on all files in project except vendored source and Hive Thrift definitions
fmt:
	find . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -w

# validates no unstaged changes exist
ci-validate: verify-codegen all-charts metering-manifests fmt
	@echo Checking for unstaged changes
	git diff --stat HEAD --ignore-submodules --exit-code

ci-validate-docker:
	docker run -i $(METERING_E2E_IMAGE):$(IMAGE_TAG) bash -c 'make ci-validate'


.PHONY: run-metering-operator-local
run-metering-operator-local: metering-operator-docker-build
	./hack/run-metering-operator-local.sh

reporting-operator-bin: $(REPORTING_OPERATOR_BIN_OUT)

reporting-operator-local: $(REPORTING_OPERATOR_GO_FILES)
	$(MAKE) build-reporting-operator REPORTING_OPERATOR_BIN_OUT=$(REPORTING_OPERATOR_BIN_OUT_LOCAL) GOOS=$(shell go env GOOS)

.PHONY: run-reporting-operator-local
run-reporting-operator-local:
	$(MAKE) reporting-operator-local
	./hack/run-local-with-port-forward.sh $(REPORTING_OPERATOR_ARGS)

$(REPORTING_OPERATOR_BIN_OUT): $(REPORTING_OPERATOR_GO_FILES)
	$(MAKE) build-reporting-operator

build-reporting-operator: $(REPORTING_OPERATOR_BIN_DEPENDENCIES) $(REPORTING_OPERATOR_GO_FILES)
	@:$(call check_defined, REPORTING_OPERATOR_BIN_OUT, Path to output binary location)
	mkdir -p $(dir $(REPORTING_OPERATOR_BIN_OUT))
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build $(GO_BUILD_ARGS) -o $(REPORTING_OPERATOR_BIN_OUT) $(REPORTING_OPERATOR_PKG)

bin/metering-override-values.yaml: ./hack/render-metering-chart-override-values.sh
	@mkdir -p bin
	./hack/render-metering-chart-override-values.sh $(RELEASE_TAG) > $@

CHART_DEPS := bin/openshift-metering-0.1.0.tgz \
	bin/operator-metering-0.1.0.tgz

all-charts: $(CHART_DEPS)

openshift-metering-chart: bin/openshift-metering-0.1.0.tgz

operator-metering-chart: bin/operator-metering-0.1.0.tgz


bin/openshift-metering-0.1.0.tgz: $(shell find charts -type f)
	@echo "Packaging openshift-metering chart dependencies"
	@mkdir -p bin && mkdir -p charts/openshift-metering/charts && hack/extract_helm_dep_repos.py charts/openshift-metering/requirements.yaml | xargs -I {} helm package --save=false -d charts/openshift-metering/charts charts/openshift-metering/{}
	helm package --save=false -d bin charts/openshift-metering

bin/operator-metering-0.1.0.tgz: $(shell find charts -type f)
	@echo "Packaging operator-metering chart dependencies"
	@mkdir -p bin && mkdir -p charts/operator-metering/charts && hack/extract_helm_dep_repos.py charts/operator-metering/requirements.yaml | xargs -I {} helm package --save=false -d charts/operator-metering/charts charts/operator-metering/{}
	helm package --save=false -d bin charts/operator-metering

metering-manifests:
	./hack/create-metering-manifests.sh $(RELEASE_TAG)

.PHONY: \
	test vendor fmt regenerate-hive-thrift thrift-gen \
	update-codegen verify-codegen \
	$(DOCKER_BUILD_TARGETS) $(DOCKER_PUSH_TARGETS) \
	$(DOCKER_TAG_TARGETS) $(DOCKER_PULL_TARGETS) \
	docker-build docker-tag docker-push \
	docker-build-all docker-tag-all docker-push-all \
	metering-test-docker \
	metering-e2e-docker-build \
	build-reporting-operator reporting-operator-bin reporting-operator-local \
	operator-metering-chart penshift-metering chart \
	bin/metering-override-values.yaml \
	metering-manifests bill-of-materials.json \
	install-kube-prometheus-helm

update-codegen: $(CODEGEN_OUTPUT_GO_FILES)
	./hack/update-codegen.sh

$(CODEGEN_OUTPUT_GO_FILES): $(CODEGEN_SOURCE_GO_FILES)

verify-codegen:
	./hack/verify-codegen.sh

# The results of these targets get vendored, but the targets exist for
# regenerating if needed.
regenerate-hive-thrift: pkg/hive/hive_thrift

# Download Hive git repo.
out/thrift.git:
	mkdir -p out
	git clone --single-branch --bare ${HIVE_REPO} $@

# Retrieve Hive thrift definition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift defintions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift thrift-gen

thrift-gen:
	thrift -gen go:package_prefix=${GO_PKG}/pkg/hive,package=hive_thrift -out pkg/hive thrift/TCLIService.thrift
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./pkg/hive/hive_thrift/...`; do rm -rf $$i; done

bill-of-materials.json: bill-of-materials.override.json
	license-bill-of-materials --override-file $(ROOT_DIR)/bill-of-materials.override.json ./... > $(ROOT_DIR)/bill-of-materials.json

kube-prometheus-helm-install:
	@echo "KUBECONFIG: $(KUBECONFIG)"
	helm ls
	helm version
	helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
	helm upgrade --install --namespace monitoring prometheus-operator coreos/prometheus-operator --wait
	# set https to false on kubelets for GKE and set the fullnameOverride for the
	# Prometheus resource so our service has a consistent name.
	helm upgrade --install --namespace monitoring kube-prometheus coreos/kube-prometheus -f hack/kube-prometheus-helm-values.yaml --wait

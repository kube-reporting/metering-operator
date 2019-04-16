SHELL := /bin/bash

ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
include build/check_de***REMOVED***ned.mk

# Package
GO_PKG := github.com/operator-framework/operator-metering
REPORTING_OPERATOR_PKG := $(GO_PKG)/cmd/reporting-operator
# these are directories/***REMOVED***les which get auto-generated or get reformated by
# gofmt
VERIFY_FILE_PATHS := cmd pkg test manifests Gopkg.lock

IMAGE_REPOSITORY = quay.io
IMAGE_ORG = openshift
DOCKER_BASE_URL = $(IMAGE_REPOSITORY)/$(IMAGE_ORG)

METERING_SRC_IMAGE_REPO=$(DOCKER_BASE_URL)/metering-src
METERING_SRC_IMAGE_TAG=latest

REPORTING_OPERATOR_IMAGE_REPO=$(DOCKER_BASE_URL)/origin-metering-reporting-operator
REPORTING_OPERATOR_IMAGE_TAG=latest
METERING_OPERATOR_IMAGE_REPO=$(DOCKER_BASE_URL)/origin-metering-helm-operator
METERING_OPERATOR_IMAGE_TAG=latest

REPORTING_OPERATOR_DOCKERFILE=Docker***REMOVED***le.reporting-operator
METERING_OPERATOR_DOCKERFILE=Docker***REMOVED***le.metering-operator

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

# Adds all the Go ***REMOVED***les in the repo as a dependency to the build-reporting-operator target
ifeq ($(CHECK_GO_FILES), true)
	REPORTING_OPERATOR_GO_FILES := $(shell ***REMOVED***nd $(ROOT_DIR) -name '*.go')
endif

# Adds the update-codegen dependency to the build-reporting-operator target
ifeq ($(RUN_UPDATE_CODEGEN), true)
	REPORTING_OPERATOR_BIN_DEPENDENCIES += update-codegen
	CODEGEN_SOURCE_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_source_***REMOVED***les.sh)
	CODEGEN_OUTPUT_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_output_***REMOVED***les.sh)
endif

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

all: fmt unit metering-manifests docker-build-all

docker-build-all: reporting-operator-docker-build metering-operator-docker-build

reporting-operator-docker-build: $(REPORTING_OPERATOR_DOCKERFILE)
	docker build -f $< -t $(REPORTING_OPERATOR_IMAGE_REPO):$(REPORTING_OPERATOR_IMAGE_TAG) $(ROOT_DIR)

metering-src-docker-build: Docker***REMOVED***le.src
	docker build -f $< -t $(METERING_SRC_IMAGE_REPO):$(METERING_SRC_IMAGE_TAG) $(ROOT_DIR)

metering-operator-docker-build: $(METERING_OPERATOR_DOCKERFILE)
	docker build -f $< -t $(METERING_OPERATOR_IMAGE_REPO):$(METERING_OPERATOR_IMAGE_TAG) $(ROOT_DIR)

# Runs gofmt on all ***REMOVED***les in project except vendored source and Hive Thrift de***REMOVED***nitions
fmt:
	***REMOVED***nd . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -w

# Update dependencies
vendor: Gopkg.toml
	dep ensure -v

test: unit

unit:
	hack/unit.sh

unit-docker: metering-src-docker-build
	docker run \
		--rm \
		-t \
		-w /go/src/github.com/operator-framework/operator-metering \
		-v $(PWD):/go/src/github.com/operator-framework/operator-metering \
		$(METERING_SRC_IMAGE_REPO):$(METERING_SRC_IMAGE_TAG) \
		make unit

integration:
	hack/integration.sh

integration-local: reporting-operator-local metering-operator-docker-build
	$(MAKE) integration DEPLOY_REPORTING_OPERATOR_LOCAL=true DEPLOY_METERING_OPERATOR_LOCAL=true

integration-docker: metering-src-docker-build
	docker run \
		--name metering-integration-docker \
		-t \
		-e METERING_NAMESPACE \
		-e METERING_OPERATOR_DEPLOY_REPO -e METERING_OPERATOR_DEPLOY_TAG \
		-e REPORTING_OPERATOR_DEPLOY_REPO -e REPORTING_OPERATOR_DEPLOY_TAG \
		-e KUBECONFIG=/kubecon***REMOVED***g \
		-e TEST_OUTPUT_PATH=/out \
		-w /go/src/github.com/operator-framework/operator-metering \
		-v $(KUBECONFIG):/kubecon***REMOVED***g \
		-v $(PWD):/go/src/github.com/operator-framework/operator-metering \
		-v /out \
		$(METERING_SRC_IMAGE_REPO):$(METERING_SRC_IMAGE_TAG) \
		make integration
	rm -rf bin/integration-docker-test-output
	docker cp metering-integration-docker:/out bin/integration-docker-test-output
	docker rm metering-integration-docker

e2e:
	hack/e2e.sh

e2e-local: reporting-operator-local metering-operator-docker-build
	$(MAKE) e2e DEPLOY_REPORTING_OPERATOR_LOCAL=true DEPLOY_METERING_OPERATOR_LOCAL=true

e2e-docker: metering-src-docker-build
	docker run \
		--name metering-e2e-docker \
		-t \
		-e METERING_NAMESPACE \
		-e METERING_OPERATOR_DEPLOY_REPO -e METERING_OPERATOR_DEPLOY_TAG \
		-e REPORTING_OPERATOR_DEPLOY_REPO -e REPORTING_OPERATOR_DEPLOY_TAG \
		-e KUBECONFIG=/kubecon***REMOVED***g \
		-e TEST_OUTPUT_PATH=/out \
		-w /go/src/github.com/operator-framework/operator-metering \
		-v $(KUBECONFIG):/kubecon***REMOVED***g \
		-v $(PWD):/go/src/github.com/operator-framework/operator-metering \
		-v /out \
		$(METERING_SRC_IMAGE_REPO):$(METERING_SRC_IMAGE_TAG) \
		make e2e
	rm -rf bin/e2e-docker-test-output
	docker cp metering-e2e-docker:/out bin/e2e-docker-test-output
	docker rm metering-e2e-docker

# validates no unstaged changes exist in $(VERIFY_FILE_PATHS)
verify: verify-codegen all-charts verify-manifests fmt
	@echo Checking for unstaged changes
	git diff --stat HEAD --ignore-submodules --exit-code -- $(VERIFY_FILE_PATHS)

verify-manifests: metering-manifests
	operator-courier verify ./manifests/deploy/openshift/olm/bundle

verify-docker: metering-src-docker-build
	docker run \
		--rm \
		-t \
		-w /go/src/github.com/operator-framework/operator-metering \
		-v $(PWD):/go/src/github.com/operator-framework/operator-metering \
		$(METERING_SRC_IMAGE_REPO):$(METERING_SRC_IMAGE_TAG) \
		make verify

.PHONY: run-metering-operator-local
run-metering-operator-local: metering-operator-docker-build
	./hack/run-metering-operator-local.sh

reporting-operator-bin: $(REPORTING_OPERATOR_BIN_OUT)

reporting-operator-local: $(REPORTING_OPERATOR_GO_FILES)
	$(MAKE) build-reporting-operator REPORTING_OPERATOR_BIN_OUT=$(REPORTING_OPERATOR_BIN_OUT_LOCAL) GOOS=$(shell go env GOOS)

.PHONY: run-reporting-operator-local
run-reporting-operator-local: reporting-operator-local
	./hack/run-reporting-operator-local.sh $(REPORTING_OPERATOR_ARGS)

$(REPORTING_OPERATOR_BIN_OUT): $(REPORTING_OPERATOR_GO_FILES)
	$(MAKE) build-reporting-operator

build-reporting-operator: $(REPORTING_OPERATOR_BIN_DEPENDENCIES) $(REPORTING_OPERATOR_GO_FILES)
	@:$(call check_de***REMOVED***ned, REPORTING_OPERATOR_BIN_OUT, Path to output binary location)
	mkdir -p $(dir $(REPORTING_OPERATOR_BIN_OUT))
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build $(GO_BUILD_ARGS) -o $(REPORTING_OPERATOR_BIN_OUT) $(REPORTING_OPERATOR_PKG)

CHART_DEPS := bin/openshift-metering-0.1.0.tgz

all-charts: $(CHART_DEPS)

openshift-metering-chart: bin/openshift-metering-0.1.0.tgz

bin/openshift-metering-0.1.0.tgz: $(shell ***REMOVED***nd charts -type f)
	@echo "Packaging openshift-metering chart dependencies"
	@mkdir -p bin && mkdir -p charts/openshift-metering/charts && hack/extract_helm_dep_repos.sh charts/openshift-metering/requirements.yaml | xargs -I {} helm package --save=false -d charts/openshift-metering/charts charts/openshift-metering/{}
	helm package --save=false -d bin charts/openshift-metering

metering-manifests:
	export \
		METERING_OPERATOR_IMAGE_REPO=$(METERING_OPERATOR_IMAGE_REPO) \
		METERING_OPERATOR_IMAGE_TAG=$(METERING_OPERATOR_IMAGE_TAG); \
	./hack/generate-metering-manifests.sh

bin/test2json: gotools/test2json/main.go
	go build -o bin/test2json gotools/test2json/main.go

.PHONY: \
	test vendor fmt verify \
	regenerate-hive-thrift thrift-gen \
	update-codegen verify-codegen \
	docker-build docker-tag docker-push \
	docker-build-all docker-tag-all docker-push-all \
	metering-test-docker \
	metering-src-docker-build \
	build-reporting-operator reporting-operator-bin reporting-operator-local \
	metering-manifests \
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

# Retrieve Hive thrift de***REMOVED***nition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift de***REMOVED***ntions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift thrift-gen

thrift-gen:
	thrift -gen go:package_pre***REMOVED***x=${GO_PKG}/pkg/hive,package=hive_thrift -out pkg/hive thrift/TCLIService.thrift
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./pkg/hive/hive_thrift/...`; do rm -rf $$i; done

kube-prometheus-helm-install:
	@echo "KUBECONFIG: $(KUBECONFIG)"
	helm ls
	helm version
	helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
	helm upgrade --install --namespace monitoring prometheus-operator coreos/prometheus-operator --wait
	# set https to false on kubelets for GKE and set the fullnameOverride for the
	# Prometheus resource so our service has a consistent name.
	helm upgrade --install --namespace monitoring kube-prometheus coreos/kube-prometheus -f hack/kube-prometheus-helm-values.yaml --wait

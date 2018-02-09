ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback
CHARGEBACK_GO_PKG := $(GO_PKG)/cmd/chargeback

DOCKER_BUILD_ARGS =
GO_BUILD_ARGS := -ldflags '-extldflags "-static"'

CHARGEBACK_HELM_OPERATOR_IMAGE := quay.io/coreos/chargeback-helm-operator
CHARGEBACK_IMAGE := quay.io/coreos/chargeback
HELM_OPERATOR_IMAGE := quay.io/coreos/helm-operator
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto
CODEGEN_IMAGE := quay.io/coreosinc/chargeback-codegen
CHARGEBACK_INTEGRATION_TESTS_IMAGE := quay.io/coreos/chargeback-integration-tests

GIT_SHA := $(shell git -C $(ROOT_DIR) rev-parse HEAD)

PULL_TAG_IMAGE_SOURCE ?= false
USE_LATEST_TAG ?= false
DOCKER_BUILD_CONTEXT = $(dir $(DOCKERFILE))
IMAGE_TAG = $(GIT_SHA)
TAG_IMAGE_SOURCE = $(IMAGE_NAME):$(GIT_SHA)

# Hive Git repository for Thrift definitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

CHARGEBACK_GO_FILES := $(shell go list -json $(CHARGEBACK_GO_PKG) | jq '.Deps[] | select(. | contains("github.com/coreos-inc/kube-chargeback"))' -r | xargs -I{} find $(GOPATH)/src/$(CHARGEBACK_GO_PKG) $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)

CODEGEN_SOURCE_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_source_files.sh)

CODEGEN_OUTPUT_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_output_files.sh)

# TODO: Add tests
all: fmt test docker-build-all

# Usage:
#	make docker-build DOCKERFILE= IMAGE_NAME=

docker-build:
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME):$(GIT_SHA) -f $(DOCKERFILE) $(DOCKER_BUILD_CONTEXT)
ifeq ($(USE_LATEST_TAG), true)
	$(MAKE) docker-tag IMAGE_TAG=latest
endif
ifdef BRANCH_TAG
	$(MAKE) docker-tag IMAGE_TAG=$(BRANCH_TAG)
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
ifeq ($(USE_LATEST_TAG), true)
	docker push $(IMAGE_NAME):latest
endif
ifdef BRANCH_TAG
	docker push $(IMAGE_NAME):$(BRANCH_TAG)
endif

DOCKER_TARGETS := \
	chargeback \
	chargeback-integration-tests \
	hadoop \
	hive \
	presto \
	helm-operator \
	chargeback-helm-operator
# These generate new make targets like chargeback-helm-operator-docker-build
# which can be invoked.
DOCKER_BUILD_TARGETS := $(addsuffix -docker-build, $(DOCKER_TARGETS))
DOCKER_PUSH_TARGETS := $(addsuffix -docker-push, $(DOCKER_TARGETS))
DOCKER_TAG_TARGETS := $(addsuffix -docker-tag, $(DOCKER_TARGETS))
DOCKER_PULL_TARGETS := $(addsuffix -docker-pull, $(DOCKER_TARGETS))

# The steps below run for each value of $(DOCKER_TARGETS) effectively, generating multiple Make targets.
# To make it easier to follow, each step will include an example after the evaluation.
# The example will be using the chargeback-helm-operator targets as it's example.
#
# The pattern/string manipulation below does the following (starting from the inner most expression):
# 1) strips -docker-push, -docker-tag, or -docker-pull from the target name ($@) giving us the non suffixed value from $(TARGETS)
# ex: chargeback-helm-operator-docker-build -> chargeback-helm-operator
# 2) Replaces - with _
# ex: chargeback-helm-operator -> chargeback_helm_operator
# 3) Uppercases letters
# ex: chargeback_helm_operator -> CHARGEBACK_HELM_OPERATOR
# 4) Appends _IMAGE
# ex: CHARGEBACK_HELM_OPERATOR -> CHARGEBACK_HELM_OPERATOR_IMAGE
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

chargeback-docker-build: images/chargeback/Dockerfile images/chargeback/bin/chargeback
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-integration-tests-docker-build: images/integration-tests/Dockerfile hack/util.sh hack/install.sh hack/uninstall.sh hack/alm-uninstall.sh hack/alm-install.sh hack/deploy.sh hack/default-env.sh
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_INTEGRATION_TESTS_IMAGE) DOCKER_BUILD_CONTEXT=$(ROOT_DIR)

chargeback-helm-operator-docker-build: images/chargeback-helm-operator/Dockerfile tectonic-chargeback-0.1.0.tgz helm-operator-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_HELM_OPERATOR_IMAGE)

helm-operator-docker-build: images/helm-operator/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HELM_OPERATOR_IMAGE) USE_LATEST_TAG=true

presto-docker-build: images/presto/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(PRESTO_IMAGE)

hadoop-docker-build: images/hadoop/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE) USE_LATEST_TAG=true

hive-docker-build: images/hive/Dockerfile hadoop-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-file --no-tests --only-code --keep k8s.io/gengo/boilerplate/*txt

test:
	go test ./pkg/...

# Runs gofmt on all files in project except vendored source and Hive Thrift definitions
fmt:
	find . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

chargeback-bin: images/chargeback/bin/chargeback

images/chargeback/bin/chargeback: $(CHARGEBACK_GO_FILES)
	$(MAKE) k8s-update-codegen
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux go build $(GO_BUILD_ARGS) -o $@ $(CHARGEBACK_GO_PKG)

tectonic-chargeback-chart: tectonic-chargeback-0.1.0.tgz

tectonic-chargeback-0.1.0.tgz: $(shell find charts -type f)
	helm dep update --skip-refresh charts/tectonic-chargeback
	helm package --save=false -d images/chargeback-helm-operator charts/tectonic-chargeback

chargeback-manifests: hack/chargeback-helm-operator-values.yaml hack/chargeback-alm-values.yaml
	./hack/create-installer-manifests.sh
	./hack/create-alm-csv-manifests.sh

release:
	test -n "$(RELEASE_VERSION)" # $$RELEASE_VERSION must be set
	@./hack/create-release.sh tectonic-chargeback-$(RELEASE_VERSION).zip


.PHONY: \
	test vendor fmt regenerate-hive-thrift \
	k8s-update-codegen k8s-verify-codegen \
	$(DOCKER_BUILD_TARGETS) $(DOCKER_PUSH_TARGETS) \
	$(DOCKER_TAG_TARGETS) $(DOCKER_PULL_TARGETS) \
	docker-build docker-tag docker-push \
	docker-build-all docker-tag-all docker-push-all \
	chargeback-bin tectonic-chargeback-chart \
	chargeback-manifests release

k8s-update-codegen: $(CODEGEN_OUTPUT_GO_FILES)
	./hack/update-codegen.sh

$(CODEGEN_OUTPUT_GO_FILES): $(CODEGEN_SOURCE_GO_FILES)

k8s-verify-codegen:
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
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_prefix=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done


ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback
CHARGEBACK_GO_PKG := $(GO_PKG)/cmd/chargeback

DOCKER_BUILD_ARGS := --no-cache
GO_BUILD_ARGS := -ldflags '-extldflags "-static"'

CHARGEBACK_ALM_INSTALL_IMAGE := quay.io/coreos/chargeback-alm-install
CHARGEBACK_IMAGE := quay.io/coreos/chargeback
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto
CODEGEN_IMAGE := quay.io/coreosinc/chargeback-codegen

GIT_SHA := $(shell git -C $(ROOT_DIR) rev-parse HEAD)
GIT_TAG := $(shell git -C $(ROOT_DIR) describe --tags --exact-match HEAD 2>/dev/null)

USE_LATEST_TAG ?= false
DOCKER_BUILD_CONTEXT = $(dir $(DOCKERFILE))

# Hive Git repository for Thrift definitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

CHARGEBACK_GO_FILES := $(shell go list -json $(CHARGEBACK_GO_PKG) | jq '.Deps[] | select(. | contains("github.com/coreos-inc/kube-chargeback"))' -r | xargs -I{} find $(GOPATH)/src/$(CHARGEBACK_GO_PKG) $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)

CODEGEN_SOURCE_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_source_files.sh)

CODEGEN_OUTPUT_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_output_files.sh)

# TODO: Add tests
all: fmt docker-build-all

docker-build-all: chargeback-docker-build presto-docker-build hive-docker-build chargeback-alm-install-docker-build

docker-push-all: chargeback-docker-push presto-docker-push hive-docker-push chargeback-alm-install-docker-push

# Usage:
#	make docker-build DOCKERFILE= IMAGE_NAME=

docker-build:
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME):$(GIT_SHA) -f $(DOCKERFILE) $(DOCKER_BUILD_CONTEXT)
ifeq ($(USE_LATEST_TAG), true)
	docker tag $(IMAGE_NAME):$(GIT_SHA) $(IMAGE_NAME):latest
endif
ifdef BRANCH_TAG
	docker tag $(IMAGE_NAME):$(GIT_SHA) $(IMAGE_NAME):$(BRANCH_TAG)
endif
ifdef GIT_TAG
	docker tag $(IMAGE_NAME):$(GIT_SHA) $(IMAGE_NAME):$(GIT_TAG)
endif

# Usage:
#	make docker-push IMAGE_NAME=

docker-push:
	docker push $(IMAGE_NAME):$(GIT_SHA)
ifeq ($(USE_LATEST_TAG), true)
	docker push $(IMAGE_NAME):latest
endif
ifdef BRANCH_TAG
	docker push $(IMAGE_NAME):$(BRANCH_TAG)
endif
ifdef GIT_TAG
	docker push $(IMAGE_NAME):$(GIT_TAG)
endif

dist: Documentation manifests examples hack/*.sh
	mkdir -p $@
	cp -r $? $@

dist.zip: dist
	zip -r $@ $?

chargeback-docker-build: images/chargeback/Dockerfile images/chargeback/bin/chargeback
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-docker-push:
	$(MAKE) docker-push IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-alm-install-docker-build: images/chargeback-alm-install/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_ALM_INSTALL_IMAGE) DOCKER_BUILD_CONTEXT=.

chargeback-alm-install-docker-push:
	$(MAKE) docker-push IMAGE_NAME=$(CHARGEBACK_ALM_INSTALL_IMAGE)

presto-docker-build: images/presto/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(PRESTO_IMAGE)

presto-docker-push:
	$(MAKE) docker-push IMAGE_NAME=$(PRESTO_IMAGE)

hadoop-docker-build: images/hadoop/Dockerfile
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE) USE_LATEST_TAG=true

hadoop-docker-push:
	$(MAKE) docker-push IMAGE_NAME=$(HADOOP_IMAGE)

hive-docker-build: images/hive/Dockerfile hadoop-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

hive-docker-push:
	$(MAKE) docker-push IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-file --no-tests --only-code --keep k8s.io/gengo/boilerplate/*txt

# Runs gofmt on all files in project except vendored source and Hive Thrift definitions
fmt:
	find . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

chargeback-bin: images/chargeback/bin/chargeback

images/chargeback/bin/chargeback: $(CHARGEBACK_GO_FILES)
	$(MAKE) k8s-update-codegen
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux go build $(GO_BUILD_ARGS) -o $@ $(CHARGEBACK_GO_PKG)

.PHONY: \
	vendor fmt regenerate-hive-thrift \
	k8s-update-codegen k8s-verify-codegen \
	chargeback-docker-build \
	presto-docker-build hive-docker-build hadoop-docker-build \
	chargeback-docker-push presto-docker-push \
	hive-docker-push hadoop-docker-push \
	docker-build docker-push \
	docker-build-all docker-push-all \
	chargeback-bin

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


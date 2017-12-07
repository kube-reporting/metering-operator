ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback
CHARGEBACK_GO_PKG := $(GO_PKG)/cmd/chargeback

DOCKER_BUILD_ARGS := --no-cache
GO_BUILD_ARGS := -ldflags '-extldflags "-static"'

CHARGEBACK_ALM_INSTALL_IMAGE := quay.io/coreos/chargeback-alm-install
CHARGEBACK_IMAGE := quay.io/coreos/chargeback
HELM_OPERATOR_IMAGE := quay.io/coreos/helm-operator
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto
CODEGEN_IMAGE := quay.io/coreosinc/chargeback-codegen

TARGETS := chargeback hadoop hive presto helm-operator chargeback-alm-install
DOCKER_BUILD_TARGETS := $(addsuf***REMOVED***x -docker-build, $(TARGETS))
DOCKER_PUSH_TARGETS := $(addsuf***REMOVED***x -docker-push, $(TARGETS))
DOCKER_IMAGE_TARGETS := $(CHARGEBACK_IMAGE) $(HADOOP_IMAGE) $(HIVE_IMAGE) $(PRESTO_IMAGE) $(CHARGEBACK_ALM_INSTALL_IMAGE)

GIT_SHA := $(shell git -C $(ROOT_DIR) rev-parse HEAD)

PULL_TAG_IMAGE_SOURCE ?= false
USE_LATEST_TAG ?= false
DOCKER_BUILD_CONTEXT = $(dir $(DOCKERFILE))
IMAGE_TAG = $(GIT_SHA)
TAG_IMAGE_SOURCE = $(IMAGE_NAME):$(GIT_SHA)

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

CHARGEBACK_GO_FILES := $(shell go list -json $(CHARGEBACK_GO_PKG) | jq '.Deps[] | select(. | contains("github.com/coreos-inc/kube-chargeback"))' -r | xargs -I{} ***REMOVED***nd $(GOPATH)/src/$(CHARGEBACK_GO_PKG) $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)

CODEGEN_SOURCE_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_source_***REMOVED***les.sh)

CODEGEN_OUTPUT_GO_FILES := $(shell $(ROOT_DIR)/hack/codegen_output_***REMOVED***les.sh)

# TODO: Add tests
all: fmt docker-build-all

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

docker-build-all: $(DOCKER_BUILD_TARGETS)

docker-push-all:
	(set -e ; $(foreach image, $(DOCKER_IMAGE_TARGETS), \
		$(MAKE) docker-push IMAGE_NAME=$(image) IMAGE_TAG=$(IMAGE_TAG); \
	))

docker-tag-all:
	(set -e ; $(foreach image, $(DOCKER_IMAGE_TARGETS), \
		$(MAKE) docker-tag IMAGE_NAME=$(image) IMAGE_TAG=$(IMAGE_TAG); \
	))

docker-pull-all:
	(set -e ; $(foreach image, $(DOCKER_IMAGE_TARGETS), \
		$(MAKE) docker-pull IMAGE_NAME=$(image) IMAGE_TAG=$(IMAGE_TAG); \
	))

dist: Documentation manifests hack
	@mkdir -p $@
	@rsync -am \
		--include hack/install.sh \
		--include hack/uninstall.sh \
		--include hack/util.sh \
		--include Documentation/Installation.md \
		--include Documentation/Report.md \
		--include Documentation/Using-chargeback.md \
		--exclude 'hack/*' \
		--exclude 'Documentation/*' \
		--exclude 'manifests/alm' \
		--exclude 'manifests/installer' \
		--exclude 'manifests/custom-resources/datastores/aws-billing.yaml' \
		--exclude manifests/chargeback/chargeback-secrets.yaml \
		$? $@

dist.zip: dist
	zip -r $@ $?

chargeback-docker-build: images/chargeback/Docker***REMOVED***le images/chargeback/bin/chargeback
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-alm-install-docker-build: images/chargeback-alm-install/Docker***REMOVED***le tectonic-chargeback-0.1.0.tgz helm-operator-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_ALM_INSTALL_IMAGE)

helm-operator-docker-build: images/helm-operator/Docker***REMOVED***le
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HELM_OPERATOR_IMAGE) USE_LATEST_TAG=true

presto-docker-build: images/presto/Docker***REMOVED***le
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(PRESTO_IMAGE)

hadoop-docker-build: images/hadoop/Docker***REMOVED***le
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE) USE_LATEST_TAG=true

hive-docker-build: images/hive/Docker***REMOVED***le hadoop-docker-build
	$(MAKE) docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code --keep k8s.io/gengo/boilerplate/*txt

# Runs gofmt on all ***REMOVED***les in project except vendored source and Hive Thrift de***REMOVED***nitions
fmt:
	***REMOVED***nd . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

chargeback-bin: images/chargeback/bin/chargeback

images/chargeback/bin/chargeback: $(CHARGEBACK_GO_FILES)
	$(MAKE) k8s-update-codegen
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux go build $(GO_BUILD_ARGS) -o $@ $(CHARGEBACK_GO_PKG)

tectonic-chargeback-chart: tectonic-chargeback-0.1.0.tgz

tectonic-chargeback-0.1.0.tgz: $(shell ***REMOVED***nd charts -type f)
	helm dep update --skip-refresh charts/tectonic-chargeback
	helm package --save=false -d images/chargeback-alm-install charts/tectonic-chargeback

.PHONY: \
	vendor fmt regenerate-hive-thrift \
	k8s-update-codegen k8s-verify-codegen \
	chargeback-docker-build \
	helm-operator-docker-build chargeback-alm-install-docker-build  \
	hadoop-docker-build presto-docker-build hive-docker-build \
	docker-build docker-tag docker-push \
	docker-build-all docker-tag-all docker-push-all \
	chargeback-bin tectonic-chargeback-chart

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

# Retrieve Hive thrift de***REMOVED***nition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift de***REMOVED***ntions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_pre***REMOVED***x=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done


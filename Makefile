ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback
CHARGEBACK_GO_PKG := $(GO_PKG)/cmd/chargeback
PROMSUM_GO_PKG := $(GO_PKG)/cmd/promsum

DOCKER_BUILD_ARGS := --no-cache
GO_BUILD_ARGS := -ldflags '-extldflags "-static"'

CHARGEBACK_IMAGE := quay.io/coreos/chargeback
PROMSUM_IMAGE := quay.io/coreos/promsum
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto

GIT_SHA := $(shell git -C $(ROOT_DIR) rev-parse HEAD)
GIT_TAG := $(shell git -C $(ROOT_DIR) describe --tags --exact-match HEAD 2>/dev/null)

USE_LATEST_TAG ?= false

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

CHARGEBACK_GO_FILES := $(shell go list -json $(CHARGEBACK_GO_PKG) | jq '.Deps[] | select(. | contains("github.com/coreos-inc/kube-chargeback"))' -r | xargs -I{} ***REMOVED***nd $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)

PROMSUM_GO_FILES := $(shell go list -json $(PROMSUM_GO_PKG) | jq '.Deps[] | select(. | contains("github.com/coreos-inc/kube-chargeback"))' -r | xargs -I{} ***REMOVED***nd $(GOPATH)/src/{} -type f -name '*.go' | sort | uniq)

# TODO: Add tests
all: fmt docker-build-all

docker-build-all: chargeback-docker-build promsum-docker-build presto-docker-build hive-docker-build

docker-push-all: chargeback-docker-push promsum-docker-push presto-docker-push hive-docker-push

# Usage:
#	make docker-build DOCKERFILE= IMAGE_NAME=

docker-build:
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME):$(GIT_SHA) -f $(DOCKERFILE) $(dir $(DOCKERFILE))
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

promsum-docker-build: images/promsum/Docker***REMOVED***le images/promsum/bin/promsum
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(PROMSUM_IMAGE)

promsum-docker-push:
	make docker-push IMAGE_NAME=$(PROMSUM_IMAGE)

chargeback-docker-build: images/chargeback/Docker***REMOVED***le images/chargeback/bin/chargeback
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-docker-push:
	make docker-push IMAGE_NAME=$(CHARGEBACK_IMAGE)

presto-docker-build: images/presto/Docker***REMOVED***le
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(PRESTO_IMAGE)

presto-docker-push:
	make docker-push IMAGE_NAME=$(PRESTO_IMAGE)

hadoop-docker-build: images/hadoop/Docker***REMOVED***le
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE) USE_LATEST_TAG=true

hadoop-docker-push:
	make docker-push IMAGE_NAME=$(HADOOP_IMAGE)

hive-docker-build: images/hive/Docker***REMOVED***le hadoop-docker-build
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

hive-docker-push:
	make docker-push IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code

# Runs gofmt on all ***REMOVED***les in project except vendored source and Hive Thrift de***REMOVED***nitions
fmt:
	***REMOVED***nd . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

images/chargeback/bin/chargeback: $(CHARGEBACK_GO_FILES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux go build $(GO_BUILD_ARGS) -o $@ $(CHARGEBACK_GO_PKG)
images/promsum/bin/promsum: $(PROMSUM_GO_FILES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux go build $(GO_BUILD_ARGS) -o $@ $(PROMSUM_GO_PKG)

.PHONY: vendor fmt regenerate-hive-thrift \
	chargeback-docker-build promsum-docker-build \
	presto-docker-build hive-docker-build hadoop-docker-build \
	chargeback-docker-push promsum-docker-push presto-docker-push \
	hive-docker-push hadoop-docker-push \
	docker-build docker-push \
	docker-build-all docker-push-all \

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


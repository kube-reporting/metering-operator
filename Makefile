ROOT_DIR:= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback

GO_BUILD_ARGS := -i -ldflags '-extldflags "-static"'

CHARGEBACK_IMAGE := quay.io/coreos/chargeback
PROMSUM_IMAGE := quay.io/coreos/promsum
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto

GIT_SHA := $(shell git -C $(ROOT_DIR) rev-parse HEAD)
GIT_TAG := $(shell git -C $(ROOT_DIR) describe --tags --exact-match HEAD 2>/dev/null)

# Hive Git repository for Thrift definitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

# TODO: Add tests
all: fmt chargeback-docker-build

docker-build-all: chargeback-docker-build promsum-docker-build presto-docker-build hive-docker-build

docker-push-all: chargeback-docker-push promsum-docker-push presto-docker-push hive-docker-push

# Usage:
#	make docker-build DOCKERFILE= IMAGE_NAME=

docker-build:
	docker build $(BUILD_ARGS) -t $(IMAGE_NAME):$(GIT_SHA) -f $(DOCKERFILE) $(dir $(DOCKERFILE))
	docker tag $(IMAGE_NAME):$(GIT_SHA) $(IMAGE_NAME):latest
ifdef GIT_TAG
	docker tag $(IMAGE_NAME):$(GIT_SHA) $(IMAGE_NAME):$(GIT_TAG)
endif

# Usage:
#	make docker-push IMAGE_NAME=

docker-push:
	docker push $(IMAGE_NAME):$(GIT_SHA)
	docker push $(IMAGE_NAME):latest
ifdef GIT_TAG
	docker push $(IMAGE_NAME):$(GIT_TAG)
endif

dist: Documentation manifests examples hack/*.sh
	mkdir -p $@
	cp -r $? $@

dist.zip: dist
	zip -r $@ $?

promsum-docker-build: images/promsum/Dockerfile images/promsum/bin/promsum
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(PROMSUM_IMAGE)

promsum-docker-push:
	make docker-push IMAGE_NAME=$(PROMSUM_IMAGE)

chargeback-docker-build: images/chargeback/Dockerfile images/chargeback/bin/chargeback
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(CHARGEBACK_IMAGE)

chargeback-docker-push:
	make docker-push IMAGE_NAME=$(CHARGEBACK_IMAGE)

presto-docker-build: images/presto/Dockerfile
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(PRESTO_IMAGE)

presto-docker-push:
	make docker-push IMAGE_NAME=$(PRESTO_IMAGE)

hadoop-docker-build: images/hadoop/Dockerfile
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(HADOOP_IMAGE)

hadoop-docker-push:
	make docker-push IMAGE_NAME=$(HADOOP_IMAGE)

hive-docker-build: images/hive/Dockerfile hadoop-docker-build
	make docker-build DOCKERFILE=$< IMAGE_NAME=$(HIVE_IMAGE)

hive-docker-push:
	make docker-push IMAGE_NAME=$(HIVE_IMAGE)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-file --no-tests --only-code

# Runs gofmt on all files in project except vendored source and Hive Thrift definitions
fmt:
	find . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

images/chargeback/bin/chargeback: cmd/chargeback
	mkdir -p $(dir $@)
	GOOS=linux go build $(GO_BUILD_ARGS) -o $@ ${GO_PKG}/$<
images/promsum/bin/promsum: cmd/promsum
	mkdir -p $(dir $@)
	GOOS=linux go build $(GO_BUILD_ARGS) -o $@ ${GO_PKG}/$<

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

# Retrieve Hive thrift definition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift defintions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_prefix=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done


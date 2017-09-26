# Package
GO_PKG := github.com/coreos-inc/kube-chargeback

CHARGEBACK_IMAGE := quay.io/coreos/chargeback
PROMSUM_IMAGE := quay.io/coreos/promsum
HADOOP_IMAGE := quay.io/coreos/chargeback-hadoop
HIVE_IMAGE := quay.io/coreos/chargeback-hive
PRESTO_IMAGE := quay.io/coreos/chargeback-presto

GIT_SHA := $(shell git rev-parse HEAD)
GIT_TAG := $(shell git describe --tags --exact-match HEAD 2>/dev/null)

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"


# TODO: Add tests
all: fmt chargeback-docker-build

docker-build: chargeback-docker-build promsum-docker-build presto-docker-build hive-docker-build

docker-push: chargeback-docker-push promsum-docker-push presto-docker-push hive-docker-push

dist: Documentation manifests examples hack/*.sh
	mkdir -p $@
	cp -r $? $@

dist.zip: dist
	zip -r $@ $?

promsum-docker-build: images/promsum/Docker***REMOVED***le images/promsum/bin/promsum
	docker build $(BUILD_ARGS) -t $(PROMSUM_IMAGE):$(GIT_SHA) $(dir $<)
	docker tag $(PROMSUM_IMAGE):$(GIT_SHA) $(PROMSUM_IMAGE):latest
ifdef GIT_TAG
	docker tag $(PROMSUM_IMAGE):$(GIT_SHA) $(PROMSUM_IMAGE):$(GIT_TAG)
endif

promsum-docker-push:
	docker push $(PROMSUM_IMAGE):$(GIT_SHA)
	docker push $(PROMSUM_IMAGE):latest
ifdef GIT_TAG
	docker push $(PROMSUM_IMAGE):$(GIT_TAG)
endif

chargeback-docker-build: images/chargeback/Docker***REMOVED***le images/chargeback/bin/chargeback
	docker build $(BUILD_ARGS) -t $(CHARGEBACK_IMAGE):$(GIT_SHA) $(dir $<)
	docker tag $(CHARGEBACK_IMAGE):$(GIT_SHA) $(CHARGEBACK_IMAGE):latest
ifdef GIT_TAG
	docker tag $(CHARGEBACK_IMAGE):$(GIT_SHA) $(CHARGEBACK_IMAGE):$(GIT_TAG)
endif

chargeback-docker-push:
	docker push $(CHARGEBACK_IMAGE):$(GIT_SHA)
	docker push $(CHARGEBACK_IMAGE):latest
ifdef GIT_TAG
	docker push $(CHARGEBACK_IMAGE):$(GIT_TAG)
endif

presto-docker-build: images/presto/Docker***REMOVED***le
	docker build $(BUILD_ARGS) -t $(PRESTO_IMAGE):$(GIT_SHA) $(dir $<)
	docker tag $(PRESTO_IMAGE):$(GIT_SHA) $(PRESTO_IMAGE):latest
ifdef GIT_TAG
	docker tag $(PRESTO_IMAGE):$(GIT_SHA) $(PRESTO_IMAGE):$(GIT_TAG)
endif

presto-docker-push:
	docker push $(PRESTO_IMAGE):$(GIT_SHA)
	docker push $(PRESTO_IMAGE):latest
ifdef GIT_TAG
	docker push $(PRESTO_IMAGE):$(GIT_TAG)
endif

hadoop-docker-build: images/hadoop/Docker***REMOVED***le
	docker build $(BUILD_ARGS) -t $(HADOOP_IMAGE):$(GIT_SHA) $(dir $<)
	docker tag $(HADOOP_IMAGE):$(GIT_SHA) $(HADOOP_IMAGE):latest
ifdef GIT_TAG
	docker tag $(HADOOP_IMAGE):$(GIT_SHA) $(HADOOP_IMAGE):$(GIT_TAG)
endif

hadoop-docker-push:
	docker push $(HADOOP_IMAGE):$(GIT_SHA)
	docker push $(HADOOP_IMAGE):latest
ifdef GIT_TAG
	docker push $(HADOOP_IMAGE):$(GIT_TAG)
endif

hive-docker-build: images/hive/Docker***REMOVED***le hadoop-docker-build
	docker build $(BUILD_ARGS) -t $(HIVE_IMAGE):$(GIT_SHA) $(dir $<)
	docker tag $(HIVE_IMAGE):$(GIT_SHA) $(HIVE_IMAGE):latest
ifdef GIT_TAG
	docker tag $(HIVE_IMAGE):$(GIT_SHA) $(HIVE_IMAGE):$(GIT_TAG)
endif

hive-docker-push:
	docker push $(HIVE_IMAGE):$(GIT_SHA)
	docker push $(HIVE_IMAGE):latest
ifdef GIT_TAG
	docker push $(HIVE_IMAGE):$(GIT_TAG)
endif

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code

# Runs gofmt on all ***REMOVED***les in project except vendored source and Hive Thrift de***REMOVED***nitions
fmt:
	***REMOVED***nd . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

images/chargeback/bin/chargeback: cmd/chargeback
	mkdir -p $(dir $@)
	GOOS=linux go build -i -o $@ ${GO_PKG}/$<
images/promsum/bin/promsum: cmd/promsum
	mkdir -p $(dir $@)
	GOOS=linux go build -i -o $@ ${GO_PKG}/$<

.PHONY: vendor fmt chargeback-docker-build promsum-docker-build presto-docker-build hive-docker-build hadoop-docker-build chargeback-docker-push promsum-docker-push presto-docker-push hive-docker-push hadoop-docker-push docker-build docker-push regenerate-hive-thrift

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


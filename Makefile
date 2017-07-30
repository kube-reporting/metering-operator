# Package
GO_PKG := github.com/coreos-inc/kube-chargeback

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

# Contains the SHA of the current base image.
BASE_IMAGE := images/base/IMAGE
BUILD_ARGS := --build-arg BASE_IMAGE=$$(cat $(BASE_IMAGE))

# TODO: Add tests
all: fmt chargeback-image

out:
	mkdir $@

promsum-image: images/promsum/IMAGE images/promsum/promsum $(BASE_IMAGE)
	docker build $(BUILD_ARGS) -t $$(cat $<) $(dir $<)

chargeback-image: images/chargeback/IMAGE images/chargeback/chargeback $(BASE_IMAGE)
	docker build $(BUILD_ARGS) -t $$(cat $<) $(dir $<)

images/base/IMAGE: images/base/Docker***REMOVED***le
	docker build --iid***REMOVED***le $@ $(dir $<)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code

# Runs gofmt on all ***REMOVED***les in project except vendored source and Hive Thrift de***REMOVED***nitions
fmt:
	***REMOVED***nd . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

images/chargeback/chargeback: cmd/chargeback pkg/hive/hive_thrift
	GOOS=linux go build -i -v -o $@ ${GO_PKG}/$<

images/promsum/promsum: cmd/promsum
	GOOS=linux go build -i -v -o $@ ${GO_PKG}/$<

# Download Hive git repo.
out/thrift.git: | out
	git clone --single-branch --bare --depth 1 ${HIVE_REPO} $@

# Retrieve Hive thrift de***REMOVED***nition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift de***REMOVED***ntions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_pre***REMOVED***x=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done

.PHONY: vendor chargeback-image promsum-image fmt

# Package
GO_PKG := github.com/coreos-inc/kube-chargeback

# Hive Git repository for Thrift definitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

# TODO: Add tests
all: fmt chargeback-image

dist: Documentation manifests examples hack/*.sh
	mkdir -p $@
	cp -r $? $@

dist.zip: dist
	zip -r $@ $?

out:
	mkdir $@

promsum-image: images/promsum/IMAGE images/promsum/bin/promsum
	docker build $(BUILD_ARGS) -t $$(cat $<) $(dir $<)

chargeback-image: images/chargeback/IMAGE images/chargeback/bin/chargeback
	docker build $(BUILD_ARGS) -t $$(cat $<) $(dir $<)

presto-image: images/presto/IMAGE
	docker build -t $$(cat $<) $(dir $<)

hive-image: images/hive/IMAGE
	docker build -t $$(cat $<) $(dir $<)

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-file --no-tests --only-code

# Runs gofmt on all files in project except vendored source and Hive Thrift definitions
fmt:
	find . -name '*.go' -not -path "./vendor/*" -not -path "./pkg/hive/hive_thrift/*" | xargs gofmt -s -w

images/chargeback/bin/chargeback: cmd/chargeback pkg/hive/hive_thrift
	mkdir -p $(dir $@)
	GOOS=linux go build -i -v -o $@ ${GO_PKG}/$<

images/promsum/bin/promsum: cmd/promsum
	mkdir -p $(dir $@)
	GOOS=linux go build -i -v -o $@ ${GO_PKG}/$<

# Download Hive git repo.
out/thrift.git: | out
	git clone --single-branch --bare ${HIVE_REPO} $@

# Retrieve Hive thrift definition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift defintions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_prefix=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done

.PHONY: vendor chargeback-image promsum-image presto-image hive-image fmt

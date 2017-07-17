# Package
GO_PKG := github.com/coreos-inc/kube-chargeback

# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "1fe8db618a7bbc09e041844021a2711c89355995"

out:
	mkdir $@

# Update dependencies
vendor: glide.yaml
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code

# Download Hive git repo.
out/thrift.git: out
	git clone --single-branch --bare --depth 1 ${HIVE_REPO} $@

# Retrieve Hive thrift de***REMOVED***nition from git repo.
thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

# Generate source from Hive thrift de***REMOVED***ntions and remove executable packages.
pkg/hive/hive_thrift: thrift/TCLIService.thrift
	thrift -gen go:package_pre***REMOVED***x=${GO_PKG}/$(dir $@),package=$(notdir $@) -out $(dir $@) $<
	for i in `go list -f '{{if eq .Name "main"}}{{ .Dir }}{{end}}' ./$@/...`; do rm -rf $$i; done

.PHONY: vendor

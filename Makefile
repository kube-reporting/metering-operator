# Hive Git repository for Thrift de***REMOVED***nitions
HIVE_REPO := "git://git.apache.org/hive.git"
HIVE_SHA := "93dd75d01867a2f9f740857aa7e961caf4f7b55d"

vendor:
	glide up --strip-vendor
	glide-vc --use-lock-***REMOVED***le --no-tests --only-code

thrift/TCLIService.thrift: out/thrift.git
	mkdir -p $(dir $@)
	git -C $< show ${HIVE_SHA}:service-rpc/if/$(notdir $@) > $@

out/thrift.git: out
	git clone --single-branch --bare --depth 1 ${HIVE_REPO} $@

out:
	mkdir $@

.PHONY: vendor

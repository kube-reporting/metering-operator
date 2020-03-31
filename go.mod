module github.com/operator-framework/operator-metering

go 1.13

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.16.0+incompatible
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/aws/aws-sdk-go v1.25.18
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/go-chi/chi v3.3.2+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/mock v1.3.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	// openshift release-4.4
	github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87 // indirect
	github.com/openshift/client-go v0.0.0-20190923180330-3b6373338c9b
	// olm 0.12.0
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190926160646-a61144936680
	github.com/prestodb/presto-go-client v0.0.0-20180328163046-568bdb2f6dbc
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/common v0.7.0
	github.com/robfig/cron v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/taozle/go-hive-driver v0.0.0-20181206100408-79951111cb07
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

require (
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/code-generator v0.0.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	// Pin to kube 1.13 because of OLM
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190918164019-21692a0861df
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/node-api => k8s.io/node-api v0.0.0-20190918163711-2299658ad911
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20190918162410-e45c26d066f2
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20190918161628-92eb3cb7496c
)

replace (
	git.apache.org/thrift.git v0.0.0-20171203172758-327ebb6c2b6d => github.com/apache/thrift v0.0.0-20171203172758-327ebb6c2b6d
	// Required by Helm
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	// indirect of OLM
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200221181648-8ce0047d664f
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190627172412-c44a8b61b9f4
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.0-pre1
	// required by operator-metering
	github.com/taozle/go-hive-driver => github.com/chancez/go-hive-driver v0.0.0-20190516203049-b3c680b33c4f
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.2
)

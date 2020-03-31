module github.com/operator-framework/operator-metering

go 1.13

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.16.0+incompatible
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/aws/aws-sdk-go v1.25.18
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/emicklei/go-restful v2.12.0+incompatible // indirect
	github.com/go-chi/chi v3.3.2+incompatible
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-openapi/swag v0.19.8 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/mock v1.3.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
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
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/tools v0.0.0-20200331202046-9d5940d49312 // indirect
	gonum.org/v1/gonum v0.7.0 // indirect
	k8s.io/api v0.18.0
	k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/code-generator v0.17.1
	k8s.io/gengo v0.0.0-20200205140755-e0e292d8aa12 // indirect
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
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
	// Pin to kube 1.16 because of OLM
	k8s.io/api => k8s.io/api v0.0.0-20190202010724-74b699b93c15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190202013456-d4288ab64945
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918162238-f783a3654da8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190202011228-6e4752048fde
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190918163234-a9c1f33e9fb9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918163108-da9fdfce26bb
	// k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190918163402-db86a8c7bb21
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190918162944-7a93a0ddadd8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190918162534-de037b596c1e
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190918162820-3b5c1246eb18
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190918164019-21692a0861df
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918162654-250a1838aa2c
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190918163543-cfa506e53441
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190918162108-227c654b2546
	k8s.io/node-api => k8s.io/node-api v0.0.0-20190918163711-2299658ad911
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190918161442-d4c9c65c82af
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20190918162410-e45c26d066f2
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20190918161628-92eb3cb7496c
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.2
)

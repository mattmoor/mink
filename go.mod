module github.com/mattmoor/mink

go 1.16

require (
	cloud.google.com/go/storage v1.11.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-containerregistry v0.6.0
	github.com/google/go-containerregistry/pkg/authn/k8schain v0.0.0-20210624211700-ce35c99b3faf // indirect
	github.com/google/ko v0.8.3
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/common v0.30.0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/tektoncd/cli v0.3.1-0.20210831053514-cec36355b4da
	github.com/tektoncd/pipeline v0.27.1
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20210825212027-de86158e7fda // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver v0.21.4 // indirect
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/caching v0.0.0-20210903131258-49416205fbf3
	knative.dev/eventing v0.25.0
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/net-http01 v0.25.0
	knative.dev/net-kourier v0.25.0
	knative.dev/networking v0.0.0-20210903132258-9d8ab8618e5f
	knative.dev/pkg v0.0.0-20210902173607-844a6bc45596
	knative.dev/serving v0.25.0
)

replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.4.1

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
)

// For ko
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.6

replace (
	k8s.io/api => k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.7
	k8s.io/client-go => k8s.io/client-go v0.20.7
	k8s.io/code-generator => k8s.io/code-generator v0.20.7
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200805222018-b89b7f3aae7b
)

// Copied from net-kourier v0.25.0
replace github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.9-0.20210217033140-668b12f5399d

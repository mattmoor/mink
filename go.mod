module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-containerregistry v0.4.1-0.20210128200529-19c2b639fab1
	github.com/google/ko v0.7.0
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/onsi/ginkgo v1.13.0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20210216180746-b93f7e1de644
	github.com/tektoncd/pipeline v0.21.1-0.20210224131343-0078cfe72cd2
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/ini.v1 v1.56.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/caching v0.0.0-20210216013537-daa996f87cb1
	knative.dev/eventing v0.21.0
	knative.dev/hack v0.0.0-20210203173706-8368e1f6eacf
	knative.dev/net-http01 v0.21.0
	knative.dev/net-kourier v0.21.0
	knative.dev/networking v0.0.0-20210216014426-94bfc013982b
	knative.dev/pkg v0.0.0-20210216013737-584933f8280b
	knative.dev/serving v0.21.0
)

replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.2.0

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
)

// For ko
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.6

replace (
	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
	k8s.io/code-generator => k8s.io/code-generator v0.19.7
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200805222018-b89b7f3aae7b
)

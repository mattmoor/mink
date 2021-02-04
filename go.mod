module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-containerregistry v0.4.1-0.20210128200529-19c2b639fab1
	github.com/google/ko v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/projectcontour/contour v1.12.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20210204133643-450bf770e8ea
	github.com/tektoncd/pipeline v0.20.1-0.20210204183043-62d8621d1bd6
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/caching v0.0.0-20210204170711-77321844ace3
	knative.dev/eventing v0.20.1-0.20210204154411-152e60848277
	knative.dev/hack v0.0.0-20210203173706-8368e1f6eacf
	knative.dev/net-contour v0.20.1-0.20210203024905-a3a8d4696d51
	knative.dev/net-http01 v0.20.1-0.20210203072306-73f4cb4dcfce
	knative.dev/networking v0.0.0-20210204170711-b61da138ba21
	knative.dev/pkg v0.0.0-20210204171111-887806985c09
	knative.dev/serving v0.20.1-0.20210204175418-2090edf6971e
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

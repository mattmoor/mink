module github.com/mattmoor/mink

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/golang/snappy v0.0.4
	github.com/google/go-containerregistry v0.6.0
	github.com/google/ko v0.8.3
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/vault/sdk v0.2.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/ryanuber/go-glob v1.0.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/sigstore/cosign v1.2.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/tektoncd/chains v0.5.1-0.20211019183434-e53b16ab0bd8
	github.com/tektoncd/cli v0.3.1-0.20211021054435-3aa43bb188a4
	github.com/tektoncd/pipeline v0.29.1-0.20211022010736-e73bfb11bc24
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/caching v0.0.0-20211019132135-6facf87d69eb
	knative.dev/eventing v0.26.1-0.20211022181727-e136cbbb2235
	knative.dev/hack v0.0.0-20211019034732-ced8ce706528
	knative.dev/net-http01 v0.26.1-0.20211020163553-bc23f49f333f
	knative.dev/net-kourier v0.26.1-0.20211020135652-410a53d883a6
	knative.dev/networking v0.0.0-20211021055311-e50e34d37d19
	knative.dev/pkg v0.0.0-20211019132235-ba2b2b1bf268
	knative.dev/serving v0.26.1-0.20211022182733-a45951406e94
)

replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.4.1

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
)

// For ko
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

replace (
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.20.2
	k8s.io/api => k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.4
	k8s.io/client-go => k8s.io/client-go v0.21.4
	k8s.io/code-generator => k8s.io/code-generator v0.21.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)

replace github.com/tektoncd/cli => github.com/mattmoor/cli v0.3.1-0.20210915213736-bc5603302c04

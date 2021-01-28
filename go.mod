module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-containerregistry v0.4.1-0.20210127165842-51f01e739161
	github.com/google/ko v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/projectcontour/contour v1.10.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20210121173339-383b37e7fd58
	github.com/tektoncd/pipeline v0.20.1-0.20210128163741-1eca890e74c9
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/caching v0.0.0-20210125050654-45e8de7ff96e
	knative.dev/eventing v0.20.1-0.20210128132430-1725902f7e39
	knative.dev/hack v0.0.0-20210120165453-8d623a0af457
	knative.dev/net-contour v0.20.1-0.20210128024030-7bff03576e1c
	knative.dev/net-http01 v0.20.1-0.20210128012731-86f758995bef
	knative.dev/networking v0.0.0-20210125050654-94433ab7f620
	knative.dev/pkg v0.0.0-20210127163530-0d31134d5f4e
	knative.dev/serving v0.20.1-0.20210128171031-fae6549b7456
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

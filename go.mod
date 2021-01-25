module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/google/go-containerregistry v0.3.0
	github.com/google/ko v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/projectcontour/contour v1.10.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20210121173339-383b37e7fd58
	github.com/tektoncd/pipeline v0.20.1-0.20210122160839-cf52c911e936
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/gengo v0.0.0-20201102161653-419f1598dd9a // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	knative.dev/caching v0.0.0-20210125050654-45e8de7ff96e
	knative.dev/eventing v0.20.1-0.20210125061954-340f07f40bda
	knative.dev/hack v0.0.0-20210120165453-8d623a0af457
	knative.dev/net-contour v0.20.1-0.20210122193054-450a79860c24
	knative.dev/net-http01 v0.20.1-0.20210121144253-04c1505c815e
	knative.dev/networking v0.0.0-20210125050654-94433ab7f620
	knative.dev/pkg v0.0.0-20210124203454-7101e9d4f6c6
	knative.dev/serving v0.20.1-0.20210125050555-39d33bf147c0
)

replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.2.0

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
)

// For ko
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)

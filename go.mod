module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/google/go-containerregistry v0.2.1
	github.com/google/ko v0.6.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/projectcontour/contour v1.10.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.1
	github.com/tektoncd/cli v0.3.1-0.20201220031706-b17647f53a60
	github.com/tektoncd/pipeline v0.19.1-0.20201217160705-55fdb95816a3
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.19.3 // indirect
	k8s.io/gengo v0.0.0-20201102161653-419f1598dd9a // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	knative.dev/caching v0.0.0-20201221023403-116224f358cd
	knative.dev/eventing v0.19.1-0.20201217221202-9643baffd747
	knative.dev/hack v0.0.0-20201214230143-4ed1ecb8db24
	knative.dev/net-contour v0.19.1-0.20201218030902-f99d5ae0e5b4
	knative.dev/net-http01 v0.19.1-0.20201217200503-6cc96eaf2654
	knative.dev/networking v0.0.0-20201221015803-71ac48b962eb
	knative.dev/pkg v0.0.0-20201218185703-e41409af6cff
	knative.dev/serving v0.19.1-0.20201218215002-782b7f2851a4
)

replace (
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.2.0

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/google/go-github/v32 => github.com/google/go-github/v32 v32.0.1-0.20200624231906-3d244d3d496e

	github.com/kubernetes-incubator/custom-metrics-apiserver => github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20190918110929-3d9be26a50eb

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20200107083527-379e7a80af0c

	github.com/tsenart/vegeta => github.com/tsenart/vegeta v1.2.1-0.20190917092155-ab06ddb56e2f
)

// For ko
replace (
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

	github.com/google/ko => github.com/google/ko v0.6.1-0.20201103214736-79beb3b01539
)

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)

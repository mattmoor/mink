module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/google/go-containerregistry v0.1.4
	github.com/google/ko v0.6.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/projectcontour/contour v1.10.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20201201122540-cbe386c85369
	github.com/tektoncd/pipeline v0.18.1-0.20201202001432-a42e5ab8fde0
	golang.org/x/net v0.0.0-20201026091529-146b70c837a4 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/tools v0.0.0-20201103190053-ac612affd56b // indirect
	google.golang.org/genproto v0.0.0-20200914193844-75d14daec038 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.19.3 // indirect
	k8s.io/gengo v0.0.0-20201102161653-419f1598dd9a // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	knative.dev/caching v0.0.0-20201202014037-fc10335afb00
	knative.dev/eventing v0.19.1-0.20201201072837-da18ee0a75f9
	knative.dev/hack v0.0.0-20201201234937-fddbf732e450
	knative.dev/net-contour v0.19.1-0.20201130160136-5716387a1c2d
	knative.dev/net-http01 v0.19.1-0.20201125174136-0d108680558b
	knative.dev/networking v0.0.0-20201202013937-4509f82009b3
	knative.dev/pkg v0.0.0-20201202014037-81712ed625fd
	knative.dev/serving v0.19.1-0.20201201235537-c8e2ccb33125
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

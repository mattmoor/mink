module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/docker/cli v0.0.0-20200303215952-eb310fca4956 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/google/go-containerregistry v0.1.4
	github.com/google/uuid v1.1.2 // indirect
	github.com/projectcontour/contour v1.9.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/tektoncd/cli v0.3.1-0.20201026154019-cb027b2293d7
	github.com/tektoncd/pipeline v0.17.1-0.20201026223119-51a02c8db934
	google.golang.org/genproto v0.0.0-20200914193844-75d14daec038 // indirect
	google.golang.org/grpc v1.32.0 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/caching v0.0.0-20201027015333-11e52bba5431
	knative.dev/eventing v0.18.1-0.20201027014033-e72a17c5a20b
	knative.dev/net-contour v0.18.1-0.20201026152541-a35be2e4cd22
	knative.dev/net-http01 v0.18.1-0.20201026032642-38d4866e37b4
	knative.dev/networking v0.0.0-20201027015433-dc85b99d4646
	knative.dev/pkg v0.0.0-20201027020033-875095c42ee0
	knative.dev/serving v0.18.1-0.20201027015133-7292a70cb5ed
	knative.dev/test-infra v0.0.0-20201026182042-46291de4ab66
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

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)

module github.com/mattmoor/mink

go 1.14

require (
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/docker/cli v0.0.0-20200303215952-eb310fca4956 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/google/go-containerregistry v0.1.2
	github.com/projectcontour/contour v1.4.1-0.20200507033955-65d52b253570
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/tektoncd/cli v0.3.1-0.20201002103652-763657b92d6e
	github.com/tektoncd/pipeline v0.16.3
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/caching v0.0.0-20200831163411-f6154e7455e2
	knative.dev/eventing v0.17.5
	knative.dev/net-contour v0.17.2
	knative.dev/net-http01 v0.17.1
	knative.dev/networking v0.0.0-20200831172815-5f2e0ad6215f
	knative.dev/pkg v0.0.0-20200831162708-14fb2347fb77
	knative.dev/serving v0.17.3
	knative.dev/test-infra v0.0.0-20200915193842-f4d4232c1f04
)

replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible

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
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)

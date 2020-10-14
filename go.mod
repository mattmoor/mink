module github.com/mattmoor/mink

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/docker/cli v0.0.0-20200303215952-eb310fca4956 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/google/go-containerregistry v0.1.3
	github.com/google/uuid v1.1.2 // indirect
	github.com/projectcontour/contour v1.9.0
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/tektoncd/cli v0.3.1-0.20201013163222-62784644bb60
	github.com/tektoncd/pipeline v0.17.1-0.20201014201822-2436f0d557d5
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/sys v0.0.0-20200915084602-288bc346aa39 // indirect
	google.golang.org/genproto v0.0.0-20200914193844-75d14daec038 // indirect
	google.golang.org/grpc v1.32.0 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/caching v0.0.0-20201013020830-6396f6b3272b
	knative.dev/eventing v0.18.1-0.20201014183131-7ee674342f3d
	knative.dev/net-contour v0.18.1-0.20201013153631-1701071790aa
	knative.dev/net-http01 v0.18.1-0.20201013143530-32a9cc61bc82
	knative.dev/networking v0.0.0-20201014172631-b3b2c758d7e6
	knative.dev/pkg v0.0.0-20201014184531-9e655c6b02f6
	knative.dev/serving v0.18.1-0.20201014211831-c1659663c272
	knative.dev/test-infra v0.0.0-20201014021030-ae3984a33f82
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
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)

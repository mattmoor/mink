module github.com/mattmoor/mink

go 1.14

require (
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/docker/cli v0.0.0-20200303215952-eb310fca4956 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/google/go-containerregistry v0.1.1
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/mattmoor/bindings v0.0.0-20200630032250-e0c4d6028efb
	github.com/projectcontour/contour v1.4.1-0.20200507033955-65d52b253570
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/tektoncd/cli v0.3.1-0.20200710173157-c226d0b80059
	github.com/tektoncd/pipeline v0.14.1-0.20200715152659-1b28720e32b8
	github.com/vaikas/postgressource v0.0.0-20200625143537-b116b1097b87
	github.com/vmware-tanzu/sources-for-knative v0.16.0
	github.com/vmware/govmomi v0.22.2 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/caching v0.0.0-20200714175930-11f3ba7a4c58
	knative.dev/eventing v0.16.1-0.20200715202741-4f2f82dccba1
	knative.dev/eventing-contrib v0.16.1-0.20200714080719-24a300fdb060
	knative.dev/net-contour v0.16.1-0.20200707232047-8bb1c2accf94
	knative.dev/net-http01 v0.16.1-0.20200707231847-195c4b274b88
	knative.dev/networking v0.0.0-20200714175930-aa7079ce334c
	knative.dev/pkg v0.0.0-20200715203233-3ba0019af6be
	knative.dev/serving v0.16.1-0.20200715211141-3bbfaf2d87d7
	knative.dev/test-infra v0.0.0-20200715185233-6964ba126fee
)

replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible

	github.com/cloudevents/sdk-go => github.com/cloudevents/sdk-go v1.2.0
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.0.0-RC2

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible
	github.com/google/go-github/v32 => github.com/google/go-github/v32 v32.0.1-0.20200624231906-3d244d3d496e

	github.com/kubernetes-incubator/custom-metrics-apiserver => github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20190918110929-3d9be26a50eb

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	github.com/spf13/cobra => github.com/chmouel/cobra v0.0.0-20200107083527-379e7a80af0c

	github.com/tsenart/vegeta => github.com/tsenart/vegeta v1.2.1-0.20190917092155-ab06ddb56e2f
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)

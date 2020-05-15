module github.com/mattmoor/mink

go 1.14

require (
	github.com/Azure/azure-sdk-for-go v41.0.0+incompatible // indirect
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/Shopify/sarama v1.26.1 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/docker/cli v0.0.0-20200303215952-eb310fca4956 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20191121125929-03554330a964 // indirect
	github.com/mattmoor/bindings v0.0.0-20200507005859-1497c853ed5a
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/projectcontour/contour v1.4.1-0.20200507033955-65d52b253570
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/tektoncd/pipeline v0.12.1-0.20200515125004-6ae926fde78f
	github.com/vaikas/postgressource v0.0.0-20200507150711-9f5b4bfdf226
	github.com/vmware-tanzu/sources-for-knative v0.14.1-0.20200507150803-6ebcec198af0
	github.com/vmware/govmomi v0.22.2 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	knative.dev/caching v0.0.0-20200513171358-85661149a26b
	knative.dev/eventing v0.14.1-0.20200515141201-4a46afff07a6
	knative.dev/eventing-contrib v0.14.1-0.20200515032200-98e505fd8247
	knative.dev/net-contour v0.14.1-0.20200514192159-4f1d7e15eaa1
	knative.dev/net-http01 v0.14.1-0.20200429235642-be6e66a4037b
	knative.dev/pkg v0.0.0-20200515002500-16d7b963416f
	knative.dev/serving v0.14.1-0.20200515044100-3d827ca18c00
	knative.dev/test-infra v0.0.0-20200514223200-ef4fd3ad398f
)

replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible

	github.com/cloudevents/sdk-go => github.com/cloudevents/sdk-go v1.2.0
	github.com/cloudevents/sdk-go/v2 => github.com/cloudevents/sdk-go/v2 v2.0.0-RC2

	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/kubernetes-incubator/custom-metrics-apiserver => github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20190918110929-3d9be26a50eb

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	github.com/tsenart/vegeta => github.com/tsenart/vegeta v1.2.1-0.20190917092155-ab06ddb56e2f

	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)

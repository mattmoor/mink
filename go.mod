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
	github.com/tektoncd/cli v0.3.1-0.20210105092911-ed9fefc25145
	github.com/tektoncd/pipeline v0.19.1-0.20210112101414-f128b41c2b87
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.19.3 // indirect
	k8s.io/gengo v0.0.0-20201102161653-419f1598dd9a // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	knative.dev/caching v0.0.0-20210107021736-1ee47505018d
	knative.dev/eventing v0.19.1-0.20210112102830-a414aee50a2b
	knative.dev/hack v0.0.0-20201214230143-4ed1ecb8db24
	knative.dev/net-contour v0.19.1-0.20210107041635-0df7ed8536a2
	knative.dev/net-http01 v0.19.1-0.20210107035636-2a526765856b
	knative.dev/networking v0.0.0-20210107024535-ecb89ced52d9
	knative.dev/pkg v0.0.0-20210107022335-51c72e24c179
	knative.dev/serving v0.19.1-0.20210111190837-560a4f678dd7
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

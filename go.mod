module github.com/mattmoor/mink

go 1.16

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/armon/go-metrics v0.3.10
	github.com/armon/go-radix v1.0.0
	github.com/docker/distribution v2.8.0+incompatible // indirect
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang/snappy v0.0.4
	github.com/google/go-containerregistry v0.8.1-0.20220219142810-1571d7fdc46e
	github.com/google/go-containerregistry/pkg/authn/kubernetes v0.0.0-20220216220642-00c59d91847c // indirect
	github.com/google/ko v0.8.3
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-hclog v1.1.0
	github.com/hashicorp/go-immutable-radix v1.3.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.3
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/go-version v1.4.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/vault/sdk v0.3.0
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/shurcooL/githubv4 v0.0.0-20191127044304-8f68eb5628d0 // indirect
	github.com/sigstore/cosign v1.5.2-0.20220222021941-3eb785b62c15
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/tektoncd/chains v0.8.1-0.20220220030910-fcc094ed96f6
	github.com/tektoncd/cli v0.3.1-0.20220216134609-4f34a14206ae
	github.com/tektoncd/pipeline v0.32.1-0.20220131230204-9b6ef48e8e35
	go.uber.org/atomic v1.9.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	knative.dev/caching v0.0.0-20220118175933-0c1cc094a7f4
	knative.dev/eventing v0.29.0
	knative.dev/hack v0.0.0-20220118141833-9b2ed8471e30
	knative.dev/net-http01 v0.29.0
	knative.dev/net-kourier v0.29.0
	knative.dev/networking v0.0.0-20220120043934-ec785540a732
	knative.dev/pkg v0.0.0-20220202132633-df430fa0dd96
	knative.dev/serving v0.29.2
)

replace (
	github.com/codegangsta/cli => github.com/urfave/cli v1.19.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
)

// Otherwise triggers and eventing pull in a conflicting antlr
replace github.com/google/cel-go => github.com/google/cel-go v0.9.0

// For ko
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7

replace (
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.20.2
	k8s.io/api => k8s.io/api v0.22.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.5
	k8s.io/client-go => k8s.io/client-go v0.22.5
	k8s.io/code-generator => k8s.io/code-generator v0.22.5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
)

replace github.com/tektoncd/cli => github.com/mattmoor/cli v0.3.1-0.20210915213736-bc5603302c04

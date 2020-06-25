// +build tools

/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	_ "knative.dev/test-infra/scripts"

	// codegen: hack/generate-knative.sh
	_ "knative.dev/pkg/hack"

	// pick up contour patch
	_ "knative.dev/net-contour/hack"

	// caching resource
	_ "knative.dev/caching/config"

	// All of the binary entrypoints from our config
	_ "github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/cmd/gcs-fetcher"
	_ "github.com/mattmoor/mink/cmd/webhook"
	_ "github.com/projectcontour/contour/cmd/contour"
	_ "github.com/tektoncd/pipeline/cmd/creds-init"
	_ "github.com/tektoncd/pipeline/cmd/entrypoint"
	_ "github.com/tektoncd/pipeline/cmd/git-init"
	_ "github.com/tektoncd/pipeline/cmd/imagedigestexporter"
	_ "github.com/tektoncd/pipeline/cmd/kubeconfigwriter"
	_ "github.com/tektoncd/pipeline/cmd/pullrequest-init"
	_ "github.com/vaikas/postgressource/cmd/receive_adapter"
	_ "github.com/vmware-tanzu/sources-for-knative/cmd/sources-for-knative-adapter"
	_ "knative.dev/eventing-contrib/github/cmd/receive_adapter"
	_ "knative.dev/eventing-contrib/kafka/source/cmd/receive_adapter"
	_ "knative.dev/eventing/cmd/apiserver_receive_adapter"
	_ "knative.dev/eventing/cmd/in_memory/channel_controller"
	_ "knative.dev/eventing/cmd/in_memory/channel_dispatcher"
	_ "knative.dev/eventing/cmd/mtbroker/filter"
	_ "knative.dev/eventing/cmd/mtbroker/ingress"
	_ "knative.dev/eventing/cmd/mtping"
	_ "knative.dev/eventing/cmd/ping"
	_ "knative.dev/serving/cmd/activator"
	_ "knative.dev/serving/cmd/autoscaler"
	_ "knative.dev/serving/cmd/default-domain"
	_ "knative.dev/serving/cmd/queue"

	// config directories
	_ "github.com/mattmoor/bindings/config"
	_ "github.com/projectcontour/contour/examples/contour"
	_ "github.com/tektoncd/pipeline/config"
	_ "github.com/vaikas/postgressource/config"
	_ "github.com/vmware-tanzu/sources-for-knative/config"
	_ "knative.dev/caching/config"
	_ "knative.dev/eventing-contrib/github/config/core"
	_ "knative.dev/eventing-contrib/kafka/source/config"
	_ "knative.dev/eventing/config/channels/in-memory-channel"
	_ "knative.dev/eventing/config/core/resources"
	_ "knative.dev/eventing/config/core/roles"
	_ "knative.dev/serving/config/core/deployments"
	_ "knative.dev/serving/config/core/resources"
	_ "knative.dev/serving/config/core/webhooks"
	_ "knative.dev/serving/config/post-install"
)

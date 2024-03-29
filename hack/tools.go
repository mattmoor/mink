//go:build tools
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
	_ "knative.dev/hack"

	// codegen: hack/generate-knative.sh
	_ "knative.dev/pkg/hack"

	// networking resources
	_ "knative.dev/networking/config"

	// net-kourier config
	_ "knative.dev/net-kourier/config"

	// For producing the `ko` image.
	_ "github.com/google/ko/cmd/ko"

	// All of the binary entrypoints from our config
	_ "github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/cmd/gcs-fetcher"
	_ "github.com/tektoncd/pipeline/cmd/entrypoint"
	_ "github.com/tektoncd/pipeline/cmd/git-init"
	_ "github.com/tektoncd/pipeline/cmd/imagedigestexporter"
	_ "github.com/tektoncd/pipeline/cmd/kubeconfigwriter"
	_ "github.com/tektoncd/pipeline/cmd/nop"
	_ "github.com/tektoncd/pipeline/cmd/pullrequest-init"
	_ "github.com/tektoncd/pipeline/cmd/workingdirinit"
	_ "knative.dev/eventing/cmd/apiserver_receive_adapter"
	_ "knative.dev/eventing/cmd/broker/filter"
	_ "knative.dev/eventing/cmd/broker/ingress"
	_ "knative.dev/eventing/cmd/in_memory/channel_controller"
	_ "knative.dev/eventing/cmd/in_memory/channel_dispatcher"
	_ "knative.dev/eventing/cmd/mtping"
	_ "knative.dev/serving/cmd/activator"
	_ "knative.dev/serving/cmd/autoscaler"
	_ "knative.dev/serving/cmd/default-domain"
	_ "knative.dev/serving/cmd/queue"

	// cosigned test script
	_ "github.com/sigstore/cosign/test"

	// config directories
	_ "github.com/tektoncd/pipeline/config"
	_ "knative.dev/caching/config"
	_ "knative.dev/eventing/config/channels/in-memory-channel"
	_ "knative.dev/eventing/config/channels/in-memory-channel/configmaps"
	_ "knative.dev/eventing/config/channels/in-memory-channel/deployments"
	_ "knative.dev/eventing/config/channels/in-memory-channel/resources"
	_ "knative.dev/eventing/config/channels/in-memory-channel/roles"
	_ "knative.dev/eventing/config/channels/in-memory-channel/webhooks"
	_ "knative.dev/eventing/config/core/deployments"
	_ "knative.dev/eventing/config/core/resources"
	_ "knative.dev/eventing/config/core/roles"
	_ "knative.dev/serving/config/core/300-resources"
	_ "knative.dev/serving/config/core/deployments"
	_ "knative.dev/serving/config/core/webhooks"
	_ "knative.dev/serving/config/post-install"
)

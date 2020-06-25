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

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	// TODO(mattmoor): Replace github binding with upstream.
	// TODO(mattmoor): Add kafka binding.

	"github.com/mattmoor/bindings/pkg/reconciler/cloudsqlbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/githubbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/slackbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/sqlbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/twitterbinding"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/pipeline/pkg/reconciler/pipelinerun"
	"github.com/tektoncd/pipeline/pkg/reconciler/taskrun"
	"github.com/vaikas/postgressource/pkg/reconciler/postgressource"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspherebinding"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource"
	github "knative.dev/eventing-contrib/github/pkg/reconciler/source"
	kafkasource "knative.dev/eventing-contrib/kafka/source/pkg/reconciler/source"
	"knative.dev/eventing/pkg/reconciler/apiserversource"
	"knative.dev/eventing/pkg/reconciler/channel"
	"knative.dev/eventing/pkg/reconciler/containersource"
	"knative.dev/eventing/pkg/reconciler/mtbroker"
	"knative.dev/eventing/pkg/reconciler/mtnamespace"
	pingsource "knative.dev/eventing/pkg/reconciler/pingsource"
	"knative.dev/eventing/pkg/reconciler/sinkbinding"
	"knative.dev/eventing/pkg/reconciler/subscription"
	"knative.dev/net-contour/pkg/reconciler/contour"
	"knative.dev/net-http01/pkg/challenger"
	"knative.dev/net-http01/pkg/reconciler/certificate"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/psbinding"
	"knative.dev/serving/pkg/reconciler/autoscaling/hpa"
	"knative.dev/serving/pkg/reconciler/configuration"
	"knative.dev/serving/pkg/reconciler/gc"
	"knative.dev/serving/pkg/reconciler/labeler"
	"knative.dev/serving/pkg/reconciler/revision"
	"knative.dev/serving/pkg/reconciler/route"
	"knative.dev/serving/pkg/reconciler/serverlessservice"
	"knative.dev/serving/pkg/reconciler/service"
)

var (
	entrypointImage = flag.String("entrypoint-image", "override-with-entrypoint:latest",
		"The container image containing our entrypoint binary.")
	nopImage = flag.String("nop-image", "tianon/true", "The container image used to stop sidecars")
	gitImage = flag.String("git-image", "override-with-git:latest",
		"The container image containing our Git binary.")
	credsImage = flag.String("creds-image", "override-with-creds:latest",
		"The container image for preparing our Build's credentials.")
	kubeconfigWriterImage = flag.String("kubeconfig-writer-image", "override-with-kubeconfig-writer:latest",
		"The container image containing our kubeconfig writer binary.")
	shellImage  = flag.String("shell-image", "busybox", "The container image containing a shell")
	gsutilImage = flag.String("gsutil-image", "google/cloud-sdk",
		"The container image containing gsutil")
	buildGCSFetcherImage = flag.String("build-gcs-fetcher-image", "gcr.io/cloud-builders/gcs-fetcher:latest",
		"The container image containing our GCS fetcher binary.")
	prImage = flag.String("pr-image", "override-with-pr:latest",
		"The container image containing our PR binary.")
	imageDigestExporterImage = flag.String("imagedigest-exporter-image", "override-with-imagedigest-exporter-image:latest",
		"The container image containing our image digest exporter binary.")
)

func main() {
	flag.Parse()
	images := pipeline.Images{
		EntrypointImage:          *entrypointImage,
		NopImage:                 *nopImage,
		GitImage:                 *gitImage,
		CredsImage:               *credsImage,
		KubeconfigWriterImage:    *kubeconfigWriterImage,
		ShellImage:               *shellImage,
		GsutilImage:              *gsutilImage,
		BuildGCSFetcherImage:     *buildGCSFetcherImage,
		PRImage:                  *prImage,
		ImageDigestExporterImage: *imageDigestExporterImage,
	}

	sbSelector := psbinding.WithSelector(psbinding.ExclusionSelector)
	if os.Getenv("SINK_BINDING_SELECTION_MODE") == "inclusion" {
		sbSelector = psbinding.WithSelector(psbinding.InclusionSelector)
	}
	vsbSelector := psbinding.WithSelector(psbinding.ExclusionSelector)
	if os.Getenv("VSPHERE_BINDING_SELECTION_MODE") == "inclusion" {
		vsbSelector = psbinding.WithSelector(psbinding.InclusionSelector)
	}

	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	chlr, err := challenger.New(ctx)
	if err != nil {
		log.Fatalf("Error creating challenger: %v", err)
	}

	// TODO(mattmoor): Support running this on a different (random?) port.
	go http.ListenAndServe(":8080", chlr)

	nop := func(ctx context.Context, b psbinding.Bindable) (context.Context, error) {
		return ctx, nil
	}

	sharedmain.WebhookMainWithConfig(ctx, "controller", sharedmain.ParseAndGetConfigOrDie(),
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,
		NewConversionController,

		// Serving resource controllers.
		configuration.NewController,
		labeler.NewController,
		revision.NewController,
		route.NewController,
		serverlessservice.NewController,
		service.NewController,
		gc.NewController,
		hpa.NewController,

		// Contour KIngress controller.
		contour.NewController,

		// Eventing source resource controllers.
		apiserversource.NewController,
		pingsource.NewController,
		containersource.NewController,

		// Messaging controllers.
		channel.NewController,
		subscription.NewController,

		// Eventing
		mtnamespace.NewController,
		mtbroker.NewController,

		// For each binding we have a controller and a binding webhook.
		sinkbinding.NewController, NewSinkBindingWebhook(sbSelector),

		// Tekton stuff
		taskrun.NewController("", images),
		pipelinerun.NewController("", images),

		// GitHubSource
		github.NewController,

		// KafkaSource
		kafkasource.NewController,

		// VMware stuff
		vspheresource.NewController,
		// For each binding we have a controller and a binding webhook.
		vspherebinding.NewController, NewVSphereBindingWebhook(vsbSelector),

		// PostgresSource
		postgressource.NewController,

		// HTTP01 Solver
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return certificate.NewController(ctx, cmw, chlr)
		},

		// Collection of mattmoor bindings that I need to upstream somewhere...
		githubbinding.NewController, NewBindingWebhook("githubbindings", githubbinding.ListAll, nop),
		slackbinding.NewController, NewBindingWebhook("slackbindings", slackbinding.ListAll, nop),
		twitterbinding.NewController, NewBindingWebhook("twitterbindings", twitterbinding.ListAll, nop),
		cloudsqlbinding.NewController, NewBindingWebhook("googlecloudsqlbindings", cloudsqlbinding.ListAll, nop),
		sqlbinding.NewController, NewBindingWebhook("sqlbindings", sqlbinding.ListAll, nop),
	)
}

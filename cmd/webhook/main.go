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
	"fmt"
	"log"
	"net/http"
	"os"

	chains "github.com/tektoncd/chains/pkg/reconciler/taskrun"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/reconciler/pipelinerun"
	"github.com/tektoncd/pipeline/pkg/reconciler/taskrun"
	"knative.dev/eventing/pkg/reconciler/apiserversource"
	"knative.dev/eventing/pkg/reconciler/broker"
	mttrigger "knative.dev/eventing/pkg/reconciler/broker/trigger"
	"knative.dev/eventing/pkg/reconciler/channel"
	"knative.dev/eventing/pkg/reconciler/containersource"
	"knative.dev/eventing/pkg/reconciler/eventtype"
	"knative.dev/eventing/pkg/reconciler/parallel"
	pingsource "knative.dev/eventing/pkg/reconciler/pingsource"
	"knative.dev/eventing/pkg/reconciler/sequence"
	"knative.dev/eventing/pkg/reconciler/sinkbinding"
	sourcecrd "knative.dev/eventing/pkg/reconciler/source/crd"
	"knative.dev/eventing/pkg/reconciler/subscription"
	"knative.dev/eventing/pkg/reconciler/sugar/namespace"
	"knative.dev/net-http01/pkg/challenger"
	"knative.dev/net-http01/pkg/ordermanager"
	"knative.dev/net-http01/pkg/reconciler/certificate"
	kingress "knative.dev/net-kourier/pkg/reconciler/ingress"
	network "knative.dev/networking/pkg"
	filteredinformerfactory "knative.dev/pkg/client/injection/kube/informers/factory/filtered"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/psbinding"
	"knative.dev/serving/pkg/reconciler/autoscaling/hpa"
	"knative.dev/serving/pkg/reconciler/configuration"
	"knative.dev/serving/pkg/reconciler/domainmapping"
	"knative.dev/serving/pkg/reconciler/gc"
	"knative.dev/serving/pkg/reconciler/labeler"
	"knative.dev/serving/pkg/reconciler/revision"
	"knative.dev/serving/pkg/reconciler/route"
	"knative.dev/serving/pkg/reconciler/serverlessservice"
	"knative.dev/serving/pkg/reconciler/service"
)

const http01ChallengePort = 8765

var (
	entrypointImage = flag.String("entrypoint-image", "override-with-entrypoint:latest",
		"The container image containing our entrypoint binary.")
	nopImage = flag.String("nop-image", "tianon/true", "The container image used to stop sidecars")
	gitImage = flag.String("git-image", "override-with-git:latest",
		"The container image containing our Git binary.")
	kubeconfigWriterImage = flag.String("kubeconfig-writer-image", "override-with-kubeconfig-writer:latest",
		"The container image containing our kubeconfig writer binary.")
	shellImage  = flag.String("shell-image", "busybox", "The container image containing a shell")
	gsutilImage = flag.String("gsutil-image", "google/cloud-sdk",
		"The container image containing gsutil")
	prImage = flag.String("pr-image", "override-with-pr:latest",
		"The container image containing our PR binary.")
	imageDigestExporterImage = flag.String("imagedigest-exporter-image", "override-with-imagedigest-exporter-image:latest",
		"The container image containing our image digest exporter binary.")
	experimentalDisableInTreeResolution = flag.Bool(disableInTreeResolutionFlag, false,
		"Disable resolution of taskrun and pipelinerun refs by the taskrun and pipelinerun reconcilers.")

	disableInTreeResolutionFlag = "experimental-disable-in-tree-resolution"
)

func main() {
	flag.StringVar(&ordermanager.Endpoint, "acme-endpoint", ordermanager.Endpoint,
		fmt.Sprintf("The ACME endpoint to use for certificate challenges. Production: %s, Staging: %s",
			ordermanager.Production, ordermanager.Staging))

	flag.Parse()

	images := pipeline.Images{
		EntrypointImage:          *entrypointImage,
		NopImage:                 *nopImage,
		GitImage:                 *gitImage,
		KubeconfigWriterImage:    *kubeconfigWriterImage,
		ShellImage:               *shellImage,
		GsutilImage:              *gsutilImage,
		PRImage:                  *prImage,
		ImageDigestExporterImage: *imageDigestExporterImage,
	}

	taskrunControllerConfig := taskrun.ControllerConfiguration{
		Images:                   images,
		DisableTaskRefResolution: *experimentalDisableInTreeResolution,
	}

	pipelinerunControllerConfig := pipelinerun.ControllerConfiguration{
		Images:                       images,
		DisablePipelineRefResolution: *experimentalDisableInTreeResolution,
	}

	sbSelector := psbinding.WithSelector(psbinding.ExclusionSelector)
	if os.Getenv("SINK_BINDING_SELECTION_MODE") == "inclusion" {
		sbSelector = psbinding.WithSelector(psbinding.InclusionSelector)
	}

	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})
	ctx = filteredinformerfactory.WithSelectors(ctx, v1beta1.ManagedByLabelKey)

	chlr, err := challenger.New(ctx)
	if err != nil {
		log.Fatal("Error creating challenger:", err)
	}

	go http.ListenAndServe(fmt.Sprint(":", http01ChallengePort), network.NewProbeHandler(chlr))

	sharedmain.WebhookMainWithConfig(ctx, "controller", injection.ParseAndGetRESTConfigOrDie(),
		certificates.NewController,
		newDefaultingAdmissionController,
		newValidationAdmissionController,
		newConfigValidationController,
		newConversionController,

		// Serving resource controllers.
		configuration.NewController,
		labeler.NewController,
		revision.NewController,
		route.NewController,
		serverlessservice.NewController,
		service.NewController,
		gc.NewController,
		hpa.NewController,
		domainmapping.NewController,

		// KIngress controller.
		kingress.NewController,

		// Eventing source resource controllers.
		apiserversource.NewController,
		pingsource.NewController,
		containersource.NewController,
		// Sources CRD
		sourcecrd.NewController,

		// Messaging controllers.
		channel.NewController,
		subscription.NewController,

		// Flows controllers.
		parallel.NewController,
		sequence.NewController,

		// Eventing
		namespace.NewController,
		broker.NewController,
		mttrigger.NewController,
		eventtype.NewController,

		// Flows
		parallel.NewController,
		sequence.NewController,

		// For each binding we have a controller and a binding webhook.
		sinkbinding.NewController, newSinkBindingWebhook(sbSelector),

		// Tekton stuff
		taskrun.NewController("", taskrunControllerConfig),
		pipelinerun.NewController("", pipelinerunControllerConfig),
		chains.NewController,

		// HTTP01 Solver
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return certificate.NewController(ctx, cmw, chlr, http01ChallengePort)
		},
	)
}

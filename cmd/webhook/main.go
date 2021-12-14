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
	secretName = flag.String("secret-name", "", "The name of the secret in the webhook's namespace that holds the public key for verification.")
)

func main() {
	flag.StringVar(&ordermanager.Endpoint, "acme-endpoint", ordermanager.Endpoint,
		fmt.Sprintf("The ACME endpoint to use for certificate challenges. Production: %s, Staging: %s",
			ordermanager.Production, ordermanager.Staging))

	opts := &pipeline.Options{}
	flag.StringVar(&opts.Images.EntrypointImage, "entrypoint-image", "", "The container image containing our entrypoint binary.")
	flag.StringVar(&opts.Images.NopImage, "nop-image", "", "The container image used to stop sidecars")
	flag.StringVar(&opts.Images.GitImage, "git-image", "", "The container image containing our Git binary.")
	flag.StringVar(&opts.Images.KubeconfigWriterImage, "kubeconfig-writer-image", "", "The container image containing our kubeconfig writer binary.")
	flag.StringVar(&opts.Images.ShellImage, "shell-image", "", "The container image containing a shell")
	flag.StringVar(&opts.Images.ShellImageWin, "shell-image-win", "", "The container image containing a windows shell")
	flag.StringVar(&opts.Images.GsutilImage, "gsutil-image", "", "The container image containing gsutil")
	flag.StringVar(&opts.Images.PRImage, "pr-image", "", "The container image containing our PR binary.")
	flag.StringVar(&opts.Images.ImageDigestExporterImage, "imagedigest-exporter-image", "", "The container image containing our image digest exporter binary.")

	flag.Parse()

	if err := opts.Images.Validate(); err != nil {
		log.Fatal(err)
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
		newCosignedValidatingWebhook,
		newCosignedMutatingWebhook,

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
		taskrun.NewController(opts),
		pipelinerun.NewController(opts),
		chains.NewController,

		// HTTP01 Solver
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return certificate.NewController(ctx, cmw, chlr, http01ChallengePort)
		},
	)
}

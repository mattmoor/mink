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
	"log"
	"net/http"

	"github.com/mattmoor/http01-solver/pkg/challenger"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"

	// The set of controllers this controller process runs.
	"github.com/mattmoor/http01-solver/pkg/reconciler/certificate"
	"knative.dev/eventing/pkg/reconciler/apiserversource"
	"knative.dev/eventing/pkg/reconciler/pingsource"
	"knative.dev/eventing/pkg/reconciler/sinkbinding"
	"knative.dev/net-contour/pkg/reconciler/contour"
	"knative.dev/serving/pkg/reconciler/autoscaling/hpa"
	"knative.dev/serving/pkg/reconciler/configuration"
	"knative.dev/serving/pkg/reconciler/gc"
	"knative.dev/serving/pkg/reconciler/labeler"
	"knative.dev/serving/pkg/reconciler/revision"
	"knative.dev/serving/pkg/reconciler/route"
	"knative.dev/serving/pkg/reconciler/serverlessservice"
	"knative.dev/serving/pkg/reconciler/service"
)

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	chlr, err := challenger.New(ctx)
	if err != nil {
		log.Fatalf("Error creating challenger: %v", err)
	}

	go http.ListenAndServe(":8080", chlr)

	sharedmain.MainWithContext(ctx, "controller",
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

		// HTTP01 Solver
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return certificate.NewController(ctx, cmw, chlr)
		},

		// Eventing source resource controllers.
		apiserversource.NewController,
		pingsource.NewController,

		// For each binding we have a controller and a binding webhook.
		sinkbinding.NewController, NewSinkBindingWebhook,
	)
}

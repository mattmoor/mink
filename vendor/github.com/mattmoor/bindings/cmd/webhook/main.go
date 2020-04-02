/*
Copyright 2019 The Knative Authors

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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/psbinding"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	"github.com/mattmoor/bindings/pkg/apis/bindings/v1alpha1"
	"github.com/mattmoor/bindings/pkg/reconciler/cloudsqlbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/githubbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/slackbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/sqlbinding"
	"github.com/mattmoor/bindings/pkg/reconciler/twitterbinding"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("GithubBinding"):         &v1alpha1.GithubBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("GoogleCloudSQLBinding"): &v1alpha1.GoogleCloudSQLBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("SQLBinding"):            &v1alpha1.SQLBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("SlackBinding"):          &v1alpha1.SlackBinding{},
	v1alpha1.SchemeGroupVersion.WithKind("TwitterBinding"):        &v1alpha1.TwitterBinding{},
}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"defaulting.webhook.bindings.mattmoor.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"validation.webhook.bindings.mattmoor.dev",

		// The path on which to serve the webhook.
		"/validation",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.bindings.mattmoor.dev",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
			metrics.ConfigMapName(): metrics.NewObservabilityConfigFromConfigMap,
		},
	)
}

func NewBindingWebhook(resource string, gla psbinding.GetListAll, wc psbinding.BindableContext) injection.ControllerConstructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		return psbinding.NewAdmissionController(ctx,
			// Name of the resource webhook.
			fmt.Sprintf("%s.webhook.bindings.mattmoor.dev", resource),

			// The path on which to serve the webhook.
			fmt.Sprintf("/%s", resource),

			// How to get all the Bindables for configuring the mutating webhook.
			gla,

			// How to setup the context prior to invoking Do/Undo.
			wc,
		)
	}
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	nop := func(ctx context.Context, b psbinding.Bindable) (context.Context, error) {
		return ctx, nil
	}

	sharedmain.MainWithContext(ctx, "webhook",
		// Our singleton certificate controller.
		certificates.NewController,

		// Our singleton webhook admission controllers
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,

		// For each binding we have a controller and a binding webhook.
		githubbinding.NewController, NewBindingWebhook("githubbindings", githubbinding.ListAll, nop),
		slackbinding.NewController, NewBindingWebhook("slackbindings", slackbinding.ListAll, nop),
		twitterbinding.NewController, NewBindingWebhook("twitterbindings", twitterbinding.ListAll, nop),
		cloudsqlbinding.NewController, NewBindingWebhook("googlecloudsqlbindings", cloudsqlbinding.ListAll, nop),
		sqlbinding.NewController, NewBindingWebhook("sqlbindings", sqlbinding.ListAll, nop),
	)
}

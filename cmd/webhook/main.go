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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/psbinding"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/conversion"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	// The set of controllers this controller process runs.
	"knative.dev/eventing/pkg/reconciler/apiserversource"
	"knative.dev/eventing/pkg/reconciler/pingsource"
	"knative.dev/eventing/pkg/reconciler/sinkbinding"
	"knative.dev/net-contour/pkg/reconciler/contour"
	"knative.dev/serving/pkg/reconciler/autoscaling/hpa"
	"knative.dev/serving/pkg/reconciler/configuration"
	"knative.dev/serving/pkg/reconciler/gc"
	"knative.dev/serving/pkg/reconciler/labeler"
	"knative.dev/serving/pkg/reconciler/nscert"
	"knative.dev/serving/pkg/reconciler/revision"
	"knative.dev/serving/pkg/reconciler/route"
	"knative.dev/serving/pkg/reconciler/serverlessservice"
	"knative.dev/serving/pkg/reconciler/service"

	// resource validation types
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	autoscalingv1alpha1 "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	net "knative.dev/serving/pkg/apis/networking/v1alpha1"
	"knative.dev/serving/pkg/apis/serving"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/apis/serving/v1beta1"

	// config validation constructors
	contourconfig "knative.dev/net-contour/pkg/reconciler/contour/config"
	metricsconfig "knative.dev/pkg/metrics"
	tracingconfig "knative.dev/pkg/tracing/config"
	defaultconfig "knative.dev/serving/pkg/apis/config"
	autoscalerconfig "knative.dev/serving/pkg/autoscaler/config"
	"knative.dev/serving/pkg/deployment"
	gcconfig "knative.dev/serving/pkg/gc"
	"knative.dev/serving/pkg/network"
	certconfig "knative.dev/serving/pkg/reconciler/certificate/config"
	domainconfig "knative.dev/serving/pkg/reconciler/route/config"
)

var ourTypes = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("Revision"):      &v1alpha1.Revision{},
	v1alpha1.SchemeGroupVersion.WithKind("Configuration"): &v1alpha1.Configuration{},
	v1alpha1.SchemeGroupVersion.WithKind("Route"):         &v1alpha1.Route{},
	v1alpha1.SchemeGroupVersion.WithKind("Service"):       &v1alpha1.Service{},
	v1beta1.SchemeGroupVersion.WithKind("Revision"):       &v1beta1.Revision{},
	v1beta1.SchemeGroupVersion.WithKind("Configuration"):  &v1beta1.Configuration{},
	v1beta1.SchemeGroupVersion.WithKind("Route"):          &v1beta1.Route{},
	v1beta1.SchemeGroupVersion.WithKind("Service"):        &v1beta1.Service{},
	v1.SchemeGroupVersion.WithKind("Revision"):            &v1.Revision{},
	v1.SchemeGroupVersion.WithKind("Configuration"):       &v1.Configuration{},
	v1.SchemeGroupVersion.WithKind("Route"):               &v1.Route{},
	v1.SchemeGroupVersion.WithKind("Service"):             &v1.Service{},

	autoscalingv1alpha1.SchemeGroupVersion.WithKind("PodAutoscaler"): &autoscalingv1alpha1.PodAutoscaler{},
	autoscalingv1alpha1.SchemeGroupVersion.WithKind("Metric"):        &autoscalingv1alpha1.Metric{},

	net.SchemeGroupVersion.WithKind("Certificate"):       &net.Certificate{},
	net.SchemeGroupVersion.WithKind("Ingress"):           &net.Ingress{},
	net.SchemeGroupVersion.WithKind("ServerlessService"): &net.ServerlessService{},

	// For group sources.knative.dev.
	sourcesv1alpha1.SchemeGroupVersion.WithKind("ApiServerSource"): &sourcesv1alpha1.ApiServerSource{},
	sourcesv1alpha1.SchemeGroupVersion.WithKind("PingSource"):      &sourcesv1alpha1.PingSource{},
	sourcesv1alpha1.SchemeGroupVersion.WithKind("SinkBinding"):     &sourcesv1alpha1.SinkBinding{},
}

func NewConversionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	var (
		v1alpha1_ = v1alpha1.SchemeGroupVersion.Version
		v1beta1_  = v1beta1.SchemeGroupVersion.Version
		v1_       = v1.SchemeGroupVersion.Version
	)

	return conversion.NewConversionController(ctx,
		// The path on which to serve the webhook
		"/resource-conversion",

		// Specify the types of custom resource definitions that should be converted
		map[schema.GroupKind]conversion.GroupKindConversion{
			v1.Kind("Service"): {
				DefinitionName: serving.ServicesResource.String(),
				HubVersion:     v1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					v1alpha1_: &v1alpha1.Service{},
					v1beta1_:  &v1beta1.Service{},
					v1_:       &v1.Service{},
				},
			},
			v1.Kind("Configuration"): {
				DefinitionName: serving.ConfigurationsResource.String(),
				HubVersion:     v1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					v1alpha1_: &v1alpha1.Configuration{},
					v1beta1_:  &v1beta1.Configuration{},
					v1_:       &v1.Configuration{},
				},
			},
			v1.Kind("Revision"): {
				DefinitionName: serving.RevisionsResource.String(),
				HubVersion:     v1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					v1alpha1_: &v1alpha1.Revision{},
					v1beta1_:  &v1beta1.Revision{},
					v1_:       &v1.Revision{},
				},
			},
			v1.Kind("Route"): {
				DefinitionName: serving.RoutesResource.String(),
				HubVersion:     v1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					v1alpha1_: &v1alpha1.Route{},
					v1beta1_:  &v1beta1.Route{},
					v1_:       &v1.Route{},
				},
			},
		},

		// A function that infuses the context passed to ConvertUp/ConvertDown/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},
	)
}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	store := defaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	store.WatchConfigs(cmw)

	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"webhook.mink.knative.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return v1.WithUpgradeViaDefaulting(store.ToContext(ctx))
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.webhook.mink.knative.dev",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate and default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			// Here is where you would infuse the context with state
			// (e.g. attach a store with configmap data)
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.mink.knative.dev",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			tracingconfig.ConfigName:         tracingconfig.NewTracingConfigFromConfigMap,
			autoscalerconfig.ConfigName:      autoscalerconfig.NewConfigFromConfigMap,
			certconfig.CertManagerConfigName: certconfig.NewCertManagerConfigFromConfigMap,
			gcconfig.ConfigName:              gcconfig.NewConfigFromConfigMapFunc(ctx),
			network.ConfigName:               network.NewConfigFromConfigMap,
			deployment.ConfigName:            deployment.NewConfigFromConfigMap,
			metrics.ConfigMapName():          metricsconfig.NewObservabilityConfigFromConfigMap,
			logging.ConfigMapName():          logging.NewConfigFromConfigMap,
			domainconfig.DomainConfigName:    domainconfig.NewDomainFromConfigMap,
			defaultconfig.DefaultsConfigName: defaultconfig.NewDefaultsConfigFromConfigMap,
			contourconfig.ContourConfigName:  contourconfig.NewContourFromConfigMap,
		},
	)
}

func NewSinkBindingWebhook(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	sbresolver := sinkbinding.WithContextFactory(ctx, func(types.NamespacedName) {})

	return psbinding.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"sinkbindings.webhook.mink.knative.dev",

		// The path on which to serve the webhook.
		"/sinkbindings",

		// How to get all the Bindables for configuring the mutating webhook.
		sinkbinding.ListAll,

		// How to setup the context prior to invoking Do/Undo.
		sbresolver,
	)
}

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

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
		nscert.NewController,

		// Contour KIngress controller.
		contour.NewController,

		// Eventing source resource controllers.
		apiserversource.NewController,
		pingsource.NewController,

		// For each binding we have a controller and a binding webhook.
		sinkbinding.NewController, NewSinkBindingWebhook,
	)
}

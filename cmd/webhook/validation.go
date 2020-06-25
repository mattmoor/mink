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

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	// config validation constructors
	contourconfig "knative.dev/net-contour/pkg/reconciler/contour/config"
	pkgleaderelection "knative.dev/pkg/leaderelection"
	metricsconfig "knative.dev/pkg/metrics"
	tracingconfig "knative.dev/pkg/tracing/config"
	defaultconfig "knative.dev/serving/pkg/apis/config"
	autoscalerconfig "knative.dev/serving/pkg/autoscaler/config"
	"knative.dev/serving/pkg/deployment"
	gcconfig "knative.dev/serving/pkg/gc"
	"knative.dev/serving/pkg/network"
	domainconfig "knative.dev/serving/pkg/reconciler/route/config"

	tkndefaultconfig "github.com/tektoncd/pipeline/pkg/apis/config"
	"github.com/tektoncd/pipeline/pkg/contexts"
	knedefaultconfig "knative.dev/eventing/pkg/apis/config"
	channeldefaultconfig "knative.dev/eventing/pkg/apis/messaging/config"
	knsdefaultconfig "knative.dev/serving/pkg/apis/config"
)

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	knsstore := knsdefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knsstore.WatchConfigs(cmw)

	knestore := knedefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knestore.WatchConfigs(cmw)

	tknstore := tkndefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	tknstore.WatchConfigs(cmw)

	channelStore := channeldefaultconfig.NewStore(logging.FromContext(ctx).Named("channel-config-store"))
	channelStore.WatchConfigs(cmw)

	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.webhook.mink.knative.dev",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate and default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return contexts.WithDefaultConfigurationName(channelStore.ToContext(tknstore.ToContext(knestore.ToContext(knsstore.ToContext(ctx)))))
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
			tracingconfig.ConfigName:          tracingconfig.NewTracingConfigFromConfigMap,
			autoscalerconfig.ConfigName:       autoscalerconfig.NewConfigFromConfigMap,
			gcconfig.ConfigName:               gcconfig.NewConfigFromConfigMapFunc(ctx),
			network.ConfigName:                network.NewConfigFromConfigMap,
			deployment.ConfigName:             deployment.NewConfigFromConfigMap,
			metrics.ConfigMapName():           metricsconfig.NewObservabilityConfigFromConfigMap,
			logging.ConfigMapName():           logging.NewConfigFromConfigMap,
			domainconfig.DomainConfigName:     domainconfig.NewDomainFromConfigMap,
			pkgleaderelection.ConfigMapName(): pkgleaderelection.NewConfigFromConfigMap,

			defaultconfig.DefaultsConfigName: func(cm *corev1.ConfigMap) (interface{}, error) {
				// Validate config-defaults for both serving and tekton.
				if _, err := tkndefaultconfig.NewDefaultsFromConfigMap(cm); err != nil {
					return nil, err
				}
				return knsdefaultconfig.NewDefaultsConfigFromConfigMap(cm)
			},
			contourconfig.ContourConfigName: contourconfig.NewContourFromConfigMap,
		},
	)
}

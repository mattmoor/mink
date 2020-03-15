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
	"knative.dev/eventing/pkg/apis/messaging"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	"knative.dev/eventing/pkg/apis/sources"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/webhook/resourcesemantics/conversion"

	"knative.dev/serving/pkg/apis/serving"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
)

func NewConversionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	var (
		v1alpha1_          = v1alpha1.SchemeGroupVersion.Version
		v1beta1_           = v1beta1.SchemeGroupVersion.Version
		v1_                = v1.SchemeGroupVersion.Version
		sourcesv1alpha1_   = sourcesv1alpha1.SchemeGroupVersion.Version
		sourcesv1alpha2_   = sourcesv1alpha2.SchemeGroupVersion.Version
		messagingv1alpha1_ = messagingv1alpha1.SchemeGroupVersion.Version
		messagingv1beta1_  = messagingv1beta1.SchemeGroupVersion.Version
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

			// Sources
			sourcesv1alpha2.Kind("ApiServerSource"): {
				DefinitionName: sources.ApiServerSourceResource.String(),
				HubVersion:     sourcesv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha1_: &sourcesv1alpha1.ApiServerSource{},
					sourcesv1alpha2_: &sourcesv1alpha2.ApiServerSource{},
				},
			},
			sourcesv1alpha2.Kind("PingSource"): {
				DefinitionName: sources.PingSourceResource.String(),
				HubVersion:     sourcesv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha1_: &sourcesv1alpha1.PingSource{},
					sourcesv1alpha2_: &sourcesv1alpha2.PingSource{},
				},
			},
			sourcesv1alpha2.Kind("SinkBinding"): {
				DefinitionName: sources.SinkBindingResource.String(),
				HubVersion:     sourcesv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha1_: &sourcesv1alpha1.SinkBinding{},
					sourcesv1alpha2_: &sourcesv1alpha2.SinkBinding{},
				},
			},

			// messaging
			messagingv1beta1.Kind("Channel"): {
				DefinitionName: messaging.ChannelsResource.String(),
				HubVersion:     messagingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					messagingv1alpha1_: &messagingv1alpha1.Channel{},
					messagingv1beta1_:  &messagingv1beta1.Channel{},
				},
			},
			messagingv1beta1.Kind("InMemoryChannel"): {
				DefinitionName: messaging.InMemoryChannelsResource.String(),
				HubVersion:     messagingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					messagingv1alpha1_: &messagingv1alpha1.InMemoryChannel{},
					messagingv1beta1_:  &messagingv1beta1.InMemoryChannel{},
				},
			},
		},

		// A function that infuses the context passed to ConvertUp/ConvertDown/SetDefaults with
		// custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},
	)
}

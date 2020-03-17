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

	tkndefaultconfig "github.com/tektoncd/pipeline/pkg/apis/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	knedefaultconfig "knative.dev/eventing/pkg/apis/config"
	"knative.dev/eventing/pkg/apis/eventing"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/apis/flows"
	flowsv1alpha1 "knative.dev/eventing/pkg/apis/flows/v1alpha1"
	flowsv1beta1 "knative.dev/eventing/pkg/apis/flows/v1beta1"
	"knative.dev/eventing/pkg/apis/messaging"
	channeldefaultconfig "knative.dev/eventing/pkg/apis/messaging/config"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	"knative.dev/eventing/pkg/apis/sources"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/webhook/resourcesemantics/conversion"
	knsdefaultconfig "knative.dev/serving/pkg/apis/config"

	"knative.dev/serving/pkg/apis/serving"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
)

func NewConversionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	knsstore := knsdefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knsstore.WatchConfigs(cmw)

	knestore := knedefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knestore.WatchConfigs(cmw)

	tknstore := tkndefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	tknstore.WatchConfigs(cmw)

	channelStore := channeldefaultconfig.NewStore(logging.FromContext(ctx).Named("channel-config-store"))
	channelStore.WatchConfigs(cmw)

	var (
		servingv1alpha1_ = v1alpha1.SchemeGroupVersion.Version
		servingv1beta1_  = v1beta1.SchemeGroupVersion.Version
		servingv1_       = v1.SchemeGroupVersion.Version

		eventingv1alpha1_  = eventingv1alpha1.SchemeGroupVersion.Version
		eventingv1beta1_   = eventingv1beta1.SchemeGroupVersion.Version
		messagingv1alpha1_ = messagingv1alpha1.SchemeGroupVersion.Version
		messagingv1beta1_  = messagingv1beta1.SchemeGroupVersion.Version
		flowsv1alpha1_     = flowsv1alpha1.SchemeGroupVersion.Version
		flowsv1beta1_      = flowsv1beta1.SchemeGroupVersion.Version
		sourcesv1alpha1_   = sourcesv1alpha1.SchemeGroupVersion.Version
		sourcesv1alpha2_   = sourcesv1alpha2.SchemeGroupVersion.Version
	)

	return conversion.NewConversionController(ctx,
		// The path on which to serve the webhook
		"/resource-conversion",

		// Specify the types of custom resource definitions that should be converted
		map[schema.GroupKind]conversion.GroupKindConversion{
			v1.Kind("Service"): {
				DefinitionName: serving.ServicesResource.String(),
				HubVersion:     servingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					servingv1alpha1_: &v1alpha1.Service{},
					servingv1beta1_:  &v1beta1.Service{},
					servingv1_:       &v1.Service{},
				},
			},
			v1.Kind("Configuration"): {
				DefinitionName: serving.ConfigurationsResource.String(),
				HubVersion:     servingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					servingv1alpha1_: &v1alpha1.Configuration{},
					servingv1beta1_:  &v1beta1.Configuration{},
					servingv1_:       &v1.Configuration{},
				},
			},
			v1.Kind("Revision"): {
				DefinitionName: serving.RevisionsResource.String(),
				HubVersion:     servingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					servingv1alpha1_: &v1alpha1.Revision{},
					servingv1beta1_:  &v1beta1.Revision{},
					servingv1_:       &v1.Revision{},
				},
			},
			v1.Kind("Route"): {
				DefinitionName: serving.RoutesResource.String(),
				HubVersion:     servingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					servingv1alpha1_: &v1alpha1.Route{},
					servingv1beta1_:  &v1beta1.Route{},
					servingv1_:       &v1.Route{},
				},
			},

			// eventing
			eventingv1beta1.Kind("Trigger"): {
				DefinitionName: eventing.TriggersResource.String(),
				HubVersion:     eventingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					eventingv1alpha1_: &eventingv1alpha1.Trigger{},
					eventingv1beta1_:  &eventingv1beta1.Trigger{},
				},
			},
			eventingv1beta1.Kind("Broker"): {
				DefinitionName: eventing.BrokersResource.String(),
				HubVersion:     eventingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					eventingv1alpha1_: &eventingv1alpha1.Broker{},
					eventingv1beta1_:  &eventingv1beta1.Broker{},
				},
			},
			eventingv1beta1.Kind("EventType"): {
				DefinitionName: eventing.EventTypesResource.String(),
				HubVersion:     eventingv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					eventingv1alpha1_: &eventingv1alpha1.EventType{},
					eventingv1beta1_:  &eventingv1beta1.EventType{},
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

			// flows
			flowsv1beta1.Kind("Sequence"): {
				DefinitionName: flows.SequenceResource.String(),
				HubVersion:     flowsv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					flowsv1alpha1_: &flowsv1alpha1.Sequence{},
					flowsv1beta1_:  &flowsv1beta1.Sequence{},
				},
			},
			flowsv1beta1.Kind("Parallel"): {
				DefinitionName: flows.ParallelResource.String(),
				HubVersion:     flowsv1alpha1_,
				Zygotes: map[string]conversion.ConvertibleObject{
					flowsv1alpha1_: &flowsv1alpha1.Parallel{},
					flowsv1beta1_:  &flowsv1beta1.Parallel{},
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
		},

		// A function that infuses the context passed to ConvertUp/ConvertDown/SetDefaults with
		// custom metadata.
		func(ctx context.Context) context.Context {
			return channelStore.ToContext(tknstore.ToContext(knestore.ToContext(knsstore.ToContext(ctx))))
		},
	)
}
